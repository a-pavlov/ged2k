package main

import (
	"fmt"

	"github.com/a-pavlov/ged2k/proto"
)

func main() {
	fmt.Println("Hello ged2k")
	var h proto.Hash = proto.Terminal
	fmt.Println(h)
}
