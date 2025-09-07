package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

func Read(file *os.File, ln uint32) []byte {
	var data []byte = make([]byte, ln)
	if _, err := io.ReadFull(file, data); err != nil {
		return []byte{}
	}
	return data
}

func ReadInt8(file *os.File) uint8 {
	return ByteToInt8(Read(file, 1)[0])
}

func ReadInt16(file *os.File) uint16 {
	return BytesToInt16Little(Read(file, 2))
}

func ReadInt32(file *os.File) uint32 {
	return BytesToIntLittle(Read(file, 4))
}

func Int8ToByte(num int8) byte {
	return byte(num)
}

func ByteToInt8(b byte) uint8 {
	return uint8(b)
}

func Int16ToLittleBytes(num uint16) []byte {
	bytesBuffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytesBuffer, num)
	return bytesBuffer
}

func BytesToInt16Little(b []byte) uint16 {
	bytesBuffer := bytes.NewBuffer(b)
	var x uint16
	_ = binary.Read(bytesBuffer, binary.LittleEndian, &x)
	return x
}

func IntToLittleBytes(num uint32) []byte {
	bytesBuffer := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytesBuffer, num)
	return bytesBuffer
}

func BytesToIntLittle(b []byte) uint32 {
	bytesBuffer := bytes.NewBuffer(b)
	var x uint32
	_ = binary.Read(bytesBuffer, binary.LittleEndian, &x)
	return x
}
