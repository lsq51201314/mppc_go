package main

import (
	"encoding/binary"
)

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
	num := mppcDecompress(source, dest, sourceLen, *destLen)
	if num > 0 && num <= *destLen {
		*destLen = num
		return 0
	}
	return -1
}

func uncompress2(dest []byte, destLen *int, source []byte, sourceLen int) int {
	num := *destLen
	*destLen = 0
	var destIndex int = 0
	var sourceIndex int = 0

	for sourceLen > 2 && num > 0 {
		num2 := int(binary.LittleEndian.Uint16(source[sourceIndex:])) & 0x7FFF
		if num2 <= 0 || num2+2 > sourceLen || num2 > 8192 {
			return -1
		}

		num3 := 0

		if (binary.LittleEndian.Uint16(source[sourceIndex:]) & 0x8000) != 0 {
			num3 = mppcDecompress(source[sourceIndex+2:sourceIndex+2+num2], dest[destIndex:], num2, num)
			if num3 <= 0 || num3 > num || num3 > 8192 {
				return -1
			}
		} else {
			num3 = num2
			if num3 > num {
				return -1
			}
			copy(dest[destIndex:], source[sourceIndex+2:sourceIndex+2+num3])
		}

		sourceIndex += num2 + 2
		sourceLen -= num2 + 2
		destIndex += num3
		num -= num3
		*destLen += num3
	}

	if sourceLen != 0 {
		return -1
	}
	return 0
}

func mppcDecompress(ibuf []byte, obuf []byte, isize int, osize int) int {
	if isize > 9217 {
		return -1
	}

	ptr := make([]byte, 16384)
	copy(ptr, ibuf)
	ibuf = ptr

	ptr2 := 0

	num := uint32(isize * 8)
	l := uint32(0)
	blen := uint32(7)
	bufIndex := 0
	currentObufIndex := 0

	for num > blen {
		num2 := fetch(ibuf, &bufIndex, &l)

		if num2 < 0x80000000 {
			if currentObufIndex >= len(obuf) {
				return -1
			}
			obuf[currentObufIndex] = byte(num2 >> 24)
			currentObufIndex++
			passbits(8, &l, &blen)
			continue
		}

		if num2 < 0xC0000000 {
			if currentObufIndex >= len(obuf) {
				return -1
			}
			obuf[currentObufIndex] = byte(((num2 >> 23) | 0x80) & 0xFF)
			currentObufIndex++
			passbits(9, &l, &blen)
			continue
		}

		var num3, num4 uint32

		if num2 >= 0xF0000000 {
			num3 = (num2 >> 22) & 0x3F
			num2 <<= 10

			if num2 < 0x80000000 {
				num4 = 3
				passbits(11, &l, &blen)
			} else if num2 < 0xC0000000 {
				num4 = 4 | ((num2 >> 28) & 3)
				passbits(14, &l, &blen)
			} else if num2 < 0xE0000000 {
				num4 = 8 | ((num2 >> 26) & 7)
				passbits(16, &l, &blen)
			} else if num2 < 0xF0000000 {
				num4 = 0x10 | ((num2 >> 24) & 0xF)
				passbits(18, &l, &blen)
			} else if num2 < 0xF8000000 {
				num4 = 0x20 | ((num2 >> 22) & 0x1F)
				passbits(20, &l, &blen)
			} else if num2 < 0xFC000000 {
				num4 = 0x40 | ((num2 >> 20) & 0x3F)
				passbits(22, &l, &blen)
			} else if num2 < 0xFE000000 {
				num4 = 0x80 | ((num2 >> 18) & 0x7F)
				passbits(24, &l, &blen)
			} else {
				passbits(10, &l, &blen)
				num2 = fetch(ibuf, &bufIndex, &l)

				if num2 < 0xFF000000 {
					num4 = 0x100 | ((num2 >> 16) & 0xFF)
					passbits(16, &l, &blen)
				} else if num2 < 0xFF400000 {
					num4 = 0x200 | ((num2 >> 14) & 0x1FF)
					passbits(18, &l, &blen)
				} else if num2 < 0xFF600000 {
					num4 = 0x400 | ((num2 >> 12) & 0x3FF)
					passbits(20, &l, &blen)
				} else if num2 < 0xFF700000 {
					num4 = 0x800 | ((num2 >> 10) & 0x7FF)
					passbits(22, &l, &blen)
				} else {
					if num2 >= 0xFF780000 {
						return -1
					}
					num4 = 0x1000 | ((num2 >> 8) & 0xFFF)
					passbits(24, &l, &blen)
				}
			}
		} else if num2 >= 0xE0000000 {
			num3 = ((num2 >> 20) & 0xFF) + 64
			num2 <<= 12

			if num2 < 0x80000000 {
				num4 = 3
				passbits(13, &l, &blen)
			} else if num2 < 0xC0000000 {
				num4 = 4 | ((num2 >> 28) & 3)
				passbits(16, &l, &blen)
			} else if num2 < 0xE0000000 {
				num4 = 8 | ((num2 >> 26) & 7)
				passbits(18, &l, &blen)
			} else if num2 < 0xF0000000 {
				num4 = 0x10 | ((num2 >> 24) & 0xF)
				passbits(20, &l, &blen)
			} else if num2 < 0xF8000000 {
				num4 = 0x20 | ((num2 >> 22) & 0x1F)
				passbits(22, &l, &blen)
			} else if num2 < 0xFC000000 {
				num4 = 0x40 | ((num2 >> 20) & 0x3F)
				passbits(24, &l, &blen)
			} else {
				passbits(12, &l, &blen)
				num2 = fetch(ibuf, &bufIndex, &l)

				if num2 < 0xFE000000 {
					num4 = 0x80 | ((num2 >> 18) & 0x7F)
					passbits(14, &l, &blen)
				} else if num2 < 0xFF000000 {
					num4 = 0x100 | ((num2 >> 16) & 0xFF)
					passbits(16, &l, &blen)
				} else if num2 < 0xFF400000 {
					num4 = 0x200 | ((num2 >> 14) & 0x1FF)
					passbits(18, &l, &blen)
				} else if num2 < 0xFF600000 {
					num4 = 0x400 | ((num2 >> 12) & 0x3FF)
					passbits(20, &l, &blen)
				} else if num2 < 0xFF700000 {
					num4 = 0x800 | ((num2 >> 10) & 0x7FF)
					passbits(22, &l, &blen)
				} else {
					if num2 >= 0xFF780000 {
						return -1
					}
					num4 = 0x1000 | ((num2 >> 8) & 0xFFF)
					passbits(24, &l, &blen)
				}
			}
		} else {
			num3 = ((num2 >> 16) & 0x1FFF) + 320
			num2 <<= 16

			if num2 < 0x80000000 {
				num4 = 3
				passbits(17, &l, &blen)
			} else if num2 < 0xC0000000 {
				num4 = 4 | ((num2 >> 28) & 3)
				passbits(20, &l, &blen)
			} else if num2 < 0xE0000000 {
				num4 = 8 | ((num2 >> 26) & 7)
				passbits(22, &l, &blen)
			} else if num2 < 0xF0000000 {
				num4 = 0x10 | ((num2 >> 24) & 0xF)
				passbits(24, &l, &blen)
			} else {
				passbits(16, &l, &blen)
				num2 = fetch(ibuf, &bufIndex, &l)

				if num2 < 0xF8000000 {
					num4 = 0x20 | ((num2 >> 22) & 0x1F)
					passbits(10, &l, &blen)
				} else if num2 < 0xFC000000 {
					num4 = 0x40 | ((num2 >> 20) & 0x3F)
					passbits(12, &l, &blen)
				} else if num2 < 0xFE000000 {
					num4 = 0x80 | ((num2 >> 18) & 0x7F)
					passbits(14, &l, &blen)
				} else if num2 < 0xFF000000 {
					num4 = 0x100 | ((num2 >> 16) & 0xFF)
					passbits(16, &l, &blen)
				} else if num2 < 0xFF400000 {
					num4 = 0x200 | ((num2 >> 14) & 0x1FF)
					passbits(18, &l, &blen)
				} else if num2 < 0xFF600000 {
					num4 = 0x400 | ((num2 >> 12) & 0x3FF)
					passbits(20, &l, &blen)
				} else if num2 < 0xFF700000 {
					num4 = 0x800 | ((num2 >> 10) & 0x7FF)
					passbits(22, &l, &blen)
				} else {
					if num2 >= 0xFF780000 {
						return -1
					}
					num4 = 0x1000 | ((num2 >> 8) & 0xFFF)
					passbits(24, &l, &blen)
				}
			}
		}

		if currentObufIndex-int(num3) < ptr2 || currentObufIndex+int(num4) > len(obuf) {
			return -1
		}

		lamecopy(obuf[currentObufIndex:], obuf[currentObufIndex-int(num3):], num4)
		currentObufIndex += int(num4)
	}

	return currentObufIndex - ptr2
}

func passbits(n uint32, l *uint32, blen *uint32) {
	*l += n
	*blen += n
}

func fetch(buf []byte, bufIndex *int, l *uint32) uint32 {
	*bufIndex += int(*l >> 3)
	*l &= 7

	val := binary.LittleEndian.Uint32(buf[*bufIndex:])
	converted := byteorder32(val)
	return converted << int(*l)
}

func lamecopy(dst []byte, src []byte, length uint32) {
	for i := uint32(0); i < length; i++ {
		dst[i] = src[i]
	}
}
