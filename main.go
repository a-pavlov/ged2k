package main

import (
	"fmt"

	"github.com/a-pavlov/ged2k/proto"
)

func main() {
	fmt.Println("Hello ged2k")

	data_1 := make([]proto.Serializable, 10)
	data_1[0] = &proto.Endpoint{}
	data_1[1] = &proto.Endpoint{}

	//a := make([]proto.Endpoint,2)
	//x0 := proto.Collection{a}

	x := proto.Collection{}
	x = append(x, &proto.Endpoint{})
	x = append(x, &proto.Endpoint{})

	data := []byte{0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x03, 0x00, 0x00, 0x00, 0x05, 0x00}
	sb2 := proto.StateBuffer{Data: data}
	var s uint32
	sb2.Read(&s)

	sb2.Read(&x)
	if sb2.Error() != nil {
		fmt.Println("Read error", sb2.Error())
	}

	test := x[0].(*proto.Endpoint)
	fmt.Println("TEST IP", test.Ip, "PORT", test.Port)
	tag_64 := []byte{proto.TAGTYPE_UINT64, 0x04, 0x00, 0x30, 0x31, 0x32, 0x33, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	t := proto.Tag{}
	sb3 := proto.StateBuffer{Data: tag_64}
	t.Get(&sb3)
	if t.IsUint64() {
		fmt.Println("Tag 64", t.AsUint64(), t.Name)
	}

	x2 := make([]byte, 0)
	x2 = test2(x2)
	fmt.Printf("X: %x\n", x2)

}

func test2(x []byte) []byte {
	x = append(x, 0x01)
	x = append(x, 0x02)
	return x
}

func receiver(c chan interface{}) {
	for x := range c {
		switch data := x.(type) {
		case uint8:
			fmt.Println("Recv", data)
		case uint32:
			fmt.Println("Recv", data)
		default:
			fmt.Println("Default")
		}
	}
}
