package main

import "encoding/binary"

const DECOMP_ERROR = -1
const MPPE_HIST_LEN = 8192

func Decompress(buffer []byte, dataSize int) []byte {
	ptr2 := make([]byte, 102400)
	sourceLen := len(buffer)
	num := dataSize
	var num2 int

	if num <= 8192 {
		num2 = uncompress(ptr2, &num, buffer, sourceLen)
	} else {
		num2 = uncompress2(ptr2, &num, buffer, sourceLen)
	}

	if num2 == -1 {
		return nil
	}

	result := make([]byte, num)
	copy(result, ptr2[:num])
	return result
}

func uncompress(dest []byte, destLen *int, source []byte, sourceLen int) int {
	len := mppcDecompress(source, dest, sourceLen, *destLen)
	if len > 0 && len <= *destLen {
		*destLen = len
		return 0
	}
	return -1
}

func uncompress2(dest []byte, destLen *int, source []byte, sourceLen int) int {
	dleft := *destLen
	*destLen = 0
	destIdx := 0
	sourceIdx := 0

	for sourceLen > 2 && dleft > 0 {
		slen := int(binary.LittleEndian.Uint16(source[sourceIdx:])) & 0x7FFF
		if slen <= 0 || slen+2 > sourceLen || slen > 8192 {
			return -1
		}

		var dlen int
		if (binary.LittleEndian.Uint16(source[sourceIdx:]) & 0x8000) != 0 {
			// 压缩数据
			dlen = mppcDecompress(source[sourceIdx+2:], dest[destIdx:], slen, dleft)
			if dlen <= 0 || dlen > dleft || dlen > 8192 {
				return -1
			}
		} else {
			// 未压缩数据
			dlen = slen
			if dlen > dleft {
				return -1
			}
			customFastMemoryCopy(dest[destIdx:], source[sourceIdx+2:], dlen)
		}

		sourceIdx += slen + 2
		sourceLen -= slen + 2
		destIdx += dlen
		dleft -= dlen
		*destLen += dlen
	}

	if sourceLen == 0 {
		return 0
	}
	return -1
}

func mppcDecompress(ibuf, obuf []byte, isize, osize int) int {
	if isize > (MPPE_HIST_LEN*9)/8+1 {
		return DECOMP_ERROR
	}

	input := make([]byte, 2*MPPE_HIST_LEN)
	copy(input, ibuf)
	ibuf = input

	obegin := 0
	oend := osize
	blenTotal := isize * 8
	l := 0
	blen := 7
	ibufIdx := 0
	obufIdx := 0

	for blenTotal > blen {
		// 获取位值
		val := fetch(ibuf, &ibufIdx, &l)

		if val < 0x80000000 {
			if obufIdx >= oend {
				return DECOMP_ERROR
			}
			obuf[obufIdx] = byte(val >> 24)
			obufIdx++
			passbits(8, &l, &blen)
			continue
		}

		if val < 0xc0000000 {
			if obufIdx >= oend {
				return DECOMP_ERROR
			}
			obuf[obufIdx] = byte(((val >> 23) | 0x80) & 0xff)
			obufIdx++
			passbits(9, &l, &blen)
			continue
		}

		var off, len uint32
		if val >= 0xf0000000 {
			off = (val >> 22) & 0x3f
			val <<= 10
			if val < 0x80000000 {
				len = 3
				passbits(11, &l, &blen)
			} else if val < 0xc0000000 {
				len = 4 | ((val >> 28) & 3)
				passbits(14, &l, &blen)
			} else if val < 0xe0000000 {
				len = 8 | ((val >> 26) & 7)
				passbits(16, &l, &blen)
			} else if val < 0xf0000000 {
				len = 16 | ((val >> 24) & 15)
				passbits(18, &l, &blen)
			} else if val < 0xf8000000 {
				len = 32 | ((val >> 22) & 0x1f)
				passbits(20, &l, &blen)
			} else if val < 0xfc000000 {
				len = 64 | ((val >> 20) & 0x3f)
				passbits(22, &l, &blen)
			} else if val < 0xfe000000 {
				len = 128 | ((val >> 18) & 0x7f)
				passbits(24, &l, &blen)
			} else {
				passbits(10, &l, &blen)
				val = fetch(ibuf, &ibufIdx, &l)
				if val < 0xff000000 {
					len = 256 | ((val >> 16) & 0xff)
					passbits(16, &l, &blen)
				} else if val < 0xff800000 {
					len = 0x200 | ((val >> 14) & 0x1ff)
					passbits(18, &l, &blen)
				} else if val < 0xffc00000 {
					len = 0x400 | ((val >> 12) & 0x3ff)
					passbits(20, &l, &blen)
				} else if val < 0xffe00000 {
					len = 0x800 | ((val >> 10) & 0x7ff)
					passbits(22, &l, &blen)
				} else if val < 0xfff00000 {
					len = 0x1000 | ((val >> 8) & 0xfff)
					passbits(24, &l, &blen)
				} else {
					return DECOMP_ERROR
				}
			}
		} else if val >= 0xe0000000 {
			off = ((val >> 20) & 0xff) + 64
			val <<= 12
			if val < 0x80000000 {
				len = 3
				passbits(13, &l, &blen)
			} else if val < 0xc0000000 {
				len = 4 | ((val >> 28) & 3)
				passbits(16, &l, &blen)
			} else if val < 0xe0000000 {
				len = 8 | ((val >> 26) & 7)
				passbits(18, &l, &blen)
			} else if val < 0xf0000000 {
				len = 16 | ((val >> 24) & 15)
				passbits(20, &l, &blen)
			} else if val < 0xf8000000 {
				len = 32 | ((val >> 22) & 0x1f)
				passbits(22, &l, &blen)
			} else if val < 0xfc000000 {
				len = 64 | ((val >> 20) & 0x3f)
				passbits(24, &l, &blen)
			} else {
				passbits(12, &l, &blen)
				val = fetch(ibuf, &ibufIdx, &l)
				if val < 0xfe000000 {
					len = 128 | ((val >> 18) & 0x7f)
					passbits(14, &l, &blen)
				} else if val < 0xff000000 {
					len = 256 | ((val >> 16) & 0xff)
					passbits(16, &l, &blen)
				} else if val < 0xff800000 {
					len = 0x200 | ((val >> 14) & 0x1ff)
					passbits(18, &l, &blen)
				} else if val < 0xffc00000 {
					len = 0x400 | ((val >> 12) & 0x3ff)
					passbits(20, &l, &blen)
				} else if val < 0xffe00000 {
					len = 0x800 | ((val >> 10) & 0x7ff)
					passbits(22, &l, &blen)
				} else if val < 0xfff00000 {
					len = 0x1000 | ((val >> 8) & 0xfff)
					passbits(24, &l, &blen)
				} else {
					return DECOMP_ERROR
				}
			}
		} else {
			off = ((val >> 16) & 0x1fff) + 320
			val <<= 16
			if val < 0x80000000 {
				len = 3
				passbits(17, &l, &blen)
			} else if val < 0xc0000000 {
				len = 4 | ((val >> 28) & 3)
				passbits(20, &l, &blen)
			} else if val < 0xe0000000 {
				len = 8 | ((val >> 26) & 7)
				passbits(22, &l, &blen)
			} else if val < 0xf0000000 {
				len = 16 | ((val >> 24) & 15)
				passbits(24, &l, &blen)
			} else {
				passbits(16, &l, &blen)
				val = fetch(ibuf, &ibufIdx, &l)
				if val < 0xf8000000 {
					len = 32 | ((val >> 22) & 0x1f)
					passbits(10, &l, &blen)
				} else if val < 0xfc000000 {
					len = 64 | ((val >> 20) & 0x3f)
					passbits(12, &l, &blen)
				} else if val < 0xfe000000 {
					len = 128 | ((val >> 18) & 0x7f)
					passbits(14, &l, &blen)
				} else if val < 0xff000000 {
					len = 256 | ((val >> 16) & 0xff)
					passbits(16, &l, &blen)
				} else if val < 0xff800000 {
					len = 0x200 | ((val >> 14) & 0x1ff)
					passbits(18, &l, &blen)
				} else if val < 0xffc00000 {
					len = 0x400 | ((val >> 12) & 0x3ff)
					passbits(20, &l, &blen)
				} else if val < 0xffe00000 {
					len = 0x800 | ((val >> 10) & 0x7ff)
					passbits(22, &l, &blen)
				} else if val < 0xfff00000 {
					len = 0x1000 | ((val >> 8) & 0xfff)
					passbits(24, &l, &blen)
				} else {
					return DECOMP_ERROR
				}
			}
		}

		// 复制数据
		if obufIdx-int(off) < obegin || obufIdx+int(len) > oend {
			return DECOMP_ERROR
		}

		lamecopy(obuf[obufIdx:], obuf[obufIdx-int(off):], int(len))
		obufIdx += int(len)
	}

	return obufIdx
}

func passbits(n int, l, blen *int) {
	*l += n
	*blen += n
}

func lamecopy(dst, src []byte, len int) {
	for i := 0; i < len; i++ {
		dst[i] = src[i]
	}
}

func customFastMemoryCopy(dest, src []byte, n int) {
	copy(dest, src[:n])
}

func fetch(buf []byte, bufIdx *int, l *int) uint32 {
	*bufIdx += *l >> 3
	*l &= 7
	if *bufIdx+4 > len(buf) {
		return 0
	}
	val := binary.LittleEndian.Uint32(buf[*bufIdx:])
	return byteorder32(val) << *l
}
