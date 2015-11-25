package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/elobuff/goamf"
	"io"
	"os"
)

func main() {
	l := 0
	f, err := os.Open("//Users/luyiyan/Projects/rtmp/bin/f.flv")
	if err != nil {
		panic(err)
	}
	// new buffer reader.
	reader := bufio.NewReader(f)
	// read flv's head.
	b := make([]byte, 9)
	io.ReadFull(reader, b)
	// print the flv head.
	fmt.Println("flv head;", b)

loop:
	// read the meta data.
	// read last tag.
	b = make([]byte, 4)
	io.ReadFull(reader, b)
	last := binary.BigEndian.Uint32(b)
	fmt.Println("last:", last)

	// read tag type.
	typ, _ := reader.ReadByte()
	fmt.Println(typ)

	// read tag's length
	b = make([]byte, 4)
	io.ReadFull(reader, b[1:])
	tagLen := binary.BigEndian.Uint32(b)
	fmt.Println(tagLen)

	// get timesteamp
	b = make([]byte, 4)
	io.ReadFull(reader, b[1:])
	b[0], _ = reader.ReadByte()
	timesteamp := binary.BigEndian.Uint32(b)
	fmt.Println("time:", timesteamp)

	b = make([]byte, 3)
	io.ReadFull(reader, b)
	fmt.Println(b)

	metaData := make([]byte, tagLen)
	io.ReadFull(reader, metaData)

	// print meta data.
	fmt.Println("==================meta data==================")
	br := bytes.NewReader(metaData)
	decode := amf.NewDecoder()
	for {
		v, err := decode.DecodeAmf0(br)
		if err != nil {
			if err != io.EOF {
				fmt.Println("get an error:", err)
			}
			break
		}
		fmt.Println("meta:!!", v)
	}
	l++
	if l > 10 {
		os.Exit(0)
	}
	goto loop
}
