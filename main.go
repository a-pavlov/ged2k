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
}
