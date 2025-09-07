package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mppc_go/utils/bytesEx"
	"mppc_go/utils/mppc"
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
	version := bytesEx.ReadInt32(elFile)
	if version != 805306720 {
		log.Fatal("文件版本错误")
	}
	//02.跳过时间戳(后续保持原样)
	bytesEx.ReadInt32(elFile)
	//03.保存头部数据，这部分数据不会改变
	orgLen := bytesEx.ReadInt32(elFile)
	bytesEx.Read(elFile, orgLen)
	//04.跳过(后续保持原样)
	bytesEx.ReadInt32(elFile)
	//05.跳过(后续保持原样)
	comLen := bytesEx.ReadInt32(elFile)
	bytesEx.Read(elFile, comLen) //computer
	bytesEx.ReadInt32(elFile)    //computerTimestamp
	//06.跳过(后续保持原样)
	bytesEx.ReadInt32(elFile)
	//07.跳过(后续保持原样)
	hardLen := bytesEx.ReadInt32(elFile)
	bytesEx.Read(elFile, hardLen) //hard
	fmt.Println(elFile.Seek(0, io.SeekCurrent))
	//***********************************************
	itemList := make([]ElItemList, 0)
	nLen := bytesEx.ReadInt32(elFile)
	for i := 0; i < int(nLen); i++ {
		id := bytesEx.ReadInt32(elFile)   //ID
		size := bytesEx.ReadInt16(elFile) //压缩长度
		itemList = append(itemList, ElItemList{
			ID:   int(id),
			Size: int(size),
		})
	}
	bytesEx.ReadInt32(elFile) //数据长度
	for i, v := range itemList {
		buf := bytesEx.Read(elFile, uint32(v.Size))
		dec := mppc.Decompress(buf, 84)
		enc := mppc.Compress(dec)
		fmt.Println(hex.EncodeToString(buf))
		fmt.Println(hex.EncodeToString(enc))
		fmt.Println(hex.EncodeToString(dec))
		fmt.Println("------------------------------------------------------")
		if hex.EncodeToString(buf) != hex.EncodeToString(enc) {
			log.Fatal(i, "  结果错误...")
		}
	}
}
