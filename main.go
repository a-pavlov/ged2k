package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello ged2k")
	reader := bufio.NewReader(os.Stdin)
	sc := ServerConnection{buffer: make([]byte, 1024)}
	s := Session{comm: make(chan string),
		register_connection:   make(chan *SessionConnection),
		unregister_connection: make(chan *SessionConnection),
		connections:           make(map[*SessionConnection]bool)}
	s.Start()

L:
	for {
		message, _ := reader.ReadString('\n')
		switch message {
		case "quit\n":
			break L
		case "start\n":
			go sc.Process()
		default:
			fmt.Print("RESP", message)
		}
	}

	s.Stop()
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
