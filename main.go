package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

type ElItemList struct {
	ID   int
	Size int
}

func main() {
	//载入文件
	elFile, err := os.Open("elements.data")
	if err != nil {
		log.Fatal(err)
	}
	defer elFile.Close()
	//01.读取标识 //60 01 00 30
	version := ReadInt32(elFile)
	if version != 805306720 {
		log.Fatal("文件版本错误")
	}
	//02.跳过时间戳(后续保持原样)
	ReadInt32(elFile)
	//03.保存头部数据，这部分数据不会改变
	orgLen := ReadInt32(elFile)
	Read(elFile, orgLen)
	//04.跳过(后续保持原样)
	ReadInt32(elFile)
	//05.跳过(后续保持原样)
	comLen := ReadInt32(elFile)
	Read(elFile, comLen) //computer
	ReadInt32(elFile)    //computerTimestamp
	//06.跳过(后续保持原样)
	ReadInt32(elFile)
	//07.跳过(后续保持原样)
	hardLen := ReadInt32(elFile)
	Read(elFile, hardLen) //hard
	fmt.Println(elFile.Seek(0, io.SeekCurrent))
	//***********************************************
	itemList := make([]ElItemList, 0)
	nLen := ReadInt32(elFile)
	for i := 0; i < int(nLen); i++ {
		id := ReadInt32(elFile)   //ID
		size := ReadInt16(elFile) //压缩长度
		itemList = append(itemList, ElItemList{
			ID:   int(id),
			Size: int(size),
		})
	}
	ReadInt32(elFile) //数据长度
	for i, v := range itemList {
		buf := Read(elFile, uint32(v.Size))
		dec := Decompress(buf, 84)
		enc := Compress(dec)
		fmt.Println(hex.EncodeToString(buf))
		fmt.Println(hex.EncodeToString(enc))
		fmt.Println(hex.EncodeToString(dec))
		fmt.Println("------------------------------------------------------")
		if hex.EncodeToString(buf) != hex.EncodeToString(enc) {
			log.Fatal(i, "  结果错误...")
		}
	}
}
