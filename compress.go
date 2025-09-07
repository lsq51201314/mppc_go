package main

import (
	"encoding/binary"
)

func Compress(buffer []byte) []byte {
	array := make([]byte, 16384)
	num := len(buffer)
	destLen := int(compressBound(uint32(num)))
	var success bool

	if num > 8192 {
		success = compress2(array, &destLen, buffer, num) == 0
	} else {
		success = compressSingle(array, &destLen, buffer, num) == 0
	}

	if !success {
		return buffer
	}
	return array[:destLen]
}

func compressBound(sourcelen uint32) uint32 {
	return sourcelen*9/8 + 1 + 2 + 3
}

func compressSingle(dest []byte, destLen *int, source []byte, sourceLen int) int {
	num := mppcCompress(source, dest, sourceLen, 0, 0)
	if num > 0 && num <= *destLen {
		*destLen = num
		return 0
	}
	return -1
}

func compress2(dest []byte, destLen *int, source []byte, sourceLen int) int {
	num := *destLen
	*destLen = 0
	var num2 uint32 = 0
	var num3 uint32 = 0

	for sourceLen > 0 && num > 2 {
		num4 := sourceLen
		if num4 > 8192 {
			num4 = 8192
		}

		num5 := mppcCompress(source, dest, num4, num2, num3+2)
		if num5 > 0 && num5 < num4 && num5 <= num-2 {
			binary.LittleEndian.PutUint16(dest[num3:], uint16(num5)|0x8000)
		} else {
			if num4 <= 0 || num4 > num-2 {
				return -1
			}
			num5 = num4
			copy(dest[num3+2:], source[num2:num2+uint32(num4)])
			binary.LittleEndian.PutUint16(dest[num3:], uint16(num5))
		}

		num2 += uint32(num4)
		sourceLen -= num4
		num3 += uint32(num5 + 2)
		num -= num5 + 2
		*destLen += num5 + 2
	}

	if sourceLen != 0 {
		return -1
	}
	return 0
}

func putbits(buf []byte, val uint32, n uint32, l *uint32, addrBuf *uint32) {
	*l += n
	shift := (32 - *l) % 32
	shifted := val << shift
	converted := byteorder32(shifted)

	if *addrBuf < uint32(len(buf)) {
		converted |= uint32(buf[*addrBuf])
	}

	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], converted)
	copy(buf[*addrBuf:], b[:])

	*addrBuf += *l >> 3
	*l &= 7
}

func putlit(buf []byte, c byte, l *uint32, addrBuf *uint32) {
	if c < 128 {
		putbits(buf, uint32(c), 8, l, addrBuf)
	} else {
		val := uint32(0x100) | (uint32(c) & 0x7F)
		putbits(buf, val, 9, l, addrBuf)
	}
}

func putoff(buf []byte, off uint32, l *uint32, addrBuf *uint32) {
	if off < 64 {
		putbits(buf, 0x3C0|off, 10, l, addrBuf)
	} else if off < 320 {
		putbits(buf, 0xE00|(off-64), 12, l, addrBuf)
	} else {
		putbits(buf, 0xC000|(off-320), 16, l, addrBuf)
	}
}

func mppcCompress(ibuf, obuf []byte, isize int, ptrIbuf, ptrObuf uint32) int {
	dictionary := make(map[uint16]int)
	num := ptrObuf
	num2 := ptrIbuf
	num3 := ptrIbuf + uint32(isize)
	num4 := ptrIbuf

	if ptrObuf < uint32(len(obuf)) {
		obuf[ptrObuf] = 0
	}
	var l uint32 = 0

	for num3-num4 > 2 {
		if int(num4)+1 >= len(ibuf) {
			break
		}
		key := binary.LittleEndian.Uint16(ibuf[num4:])

		var num5 int
		exists := false
		if num5, exists = dictionary[key]; !exists {
			num5 = -1
			dictionary[key] = num5
		}
		dictionary[key] = int(num4)

		if num5 < int(num2) || num5 >= int(num4) {
			putlit(obuf, ibuf[ptrIbuf], &l, &ptrObuf)
			ptrIbuf++
			num4 = ptrIbuf
			continue
		}

		if num5+1 >= len(ibuf) {
			putlit(obuf, ibuf[ptrIbuf], &l, &ptrObuf)
			ptrIbuf++
			num4++
			continue
		}
		currentNum4 := num4
		num4++
		if binary.LittleEndian.Uint16(ibuf[num5:]) != binary.LittleEndian.Uint16(ibuf[currentNum4:]) {
			putlit(obuf, ibuf[ptrIbuf], &l, &ptrObuf)
			ptrIbuf++
			continue
		}

		num5 += 2
		if num4 >= num3 || num5 >= len(ibuf) {
			putlit(obuf, ibuf[ptrIbuf], &l, &ptrObuf)
			ptrIbuf++
			num4 = ptrIbuf
			continue
		}
		num4++
		if ibuf[num5] != ibuf[num4] {
			putlit(obuf, ibuf[ptrIbuf], &l, &ptrObuf)
			ptrIbuf++
			num4 = ptrIbuf
			continue
		}

		num5++
		num4++
		for num4 < num3 && num5 < len(ibuf) && ibuf[num5] == ibuf[num4] {
			num4++
			num5++
		}

		matchLen := num4 - ptrIbuf
		ptrIbuf = num4
		putoff(obuf, uint32(num4)-uint32(num5), &l, &ptrObuf)

		switch {
		case matchLen < 4:
			putbits(obuf, 0, 1, &l, &ptrObuf)
		case matchLen < 8:
			putbits(obuf, 8|(matchLen&3), 4, &l, &ptrObuf)
		case matchLen < 16:
			putbits(obuf, 0x30|(matchLen&7), 6, &l, &ptrObuf)
		case matchLen < 32:
			putbits(obuf, 0xE0|(matchLen&0xF), 8, &l, &ptrObuf)
		case matchLen < 64:
			putbits(obuf, 0x3C0|(matchLen&0x1F), 10, &l, &ptrObuf)
		case matchLen < 128:
			putbits(obuf, 0xF80|(matchLen&0x3F), 12, &l, &ptrObuf)
		case matchLen < 256:
			putbits(obuf, 0x3F00|(matchLen&0x7F), 14, &l, &ptrObuf)
		case matchLen < 512:
			putbits(obuf, 0xFE00|(matchLen&0xFF), 16, &l, &ptrObuf)
		case matchLen < 1024:
			putbits(obuf, 0x3FC00|(matchLen&0x1FF), 18, &l, &ptrObuf)
		case matchLen < 2048:
			putbits(obuf, 0xFF800|(matchLen&0x3FF), 20, &l, &ptrObuf)
		case matchLen < 4096:
			putbits(obuf, 0x3FF000|(matchLen&0x7FF), 22, &l, &ptrObuf)
		case matchLen < 8192:
			putbits(obuf, 0xFFE000|(matchLen&0xFFF), 24, &l, &ptrObuf)
		}
	}

	switch num3 - num4 {
	case 2:
		putlit(obuf, ibuf[ptrIbuf], &l, &ptrObuf)
		ptrIbuf++
		putlit(obuf, ibuf[ptrIbuf], &l, &ptrObuf)
		ptrIbuf++
	case 1:
		putlit(obuf, ibuf[ptrIbuf], &l, &ptrObuf)
		ptrIbuf++
	}

	if l != 0 {
		putbits(obuf, 0, 8-l, &l, &ptrObuf)
	}

	return int(ptrObuf - num)
}

func byteorder32(x uint32) uint32 {
	return (x&0x000000FF)<<24 |
		(x&0x0000FF00)<<8 |
		(x&0x00FF0000)>>8 |
		(x&0xFF000000)>>24
}
