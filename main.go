package main

import (
	"fmt"

	"github.com/a-pavlov/ged2k/proto"
)

func main() {
	fmt.Println("Hello ged2k")
	var h proto.Hash = proto.Terminal
	fmt.Println(h)

	tag1 := proto.Tag{}
	buf := []byte{proto.TAGTYPE_UINT16 | 0x80, 0x11, 0x0A, 0x0D}
	sb := proto.StateBuffer{Data: buf}
	tag1.Get(&sb)
	//uint16(tag1.Value)
}
