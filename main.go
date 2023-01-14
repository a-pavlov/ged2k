package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Hello ged2k")
	reader := bufio.NewReader(os.Stdin)
	cfg := Config{Port: 30000, Name: "TestGed2k"}
	s := CreateSession(cfg)
	s.Start()

L:
	for {
		message, _ := reader.ReadString('\n')
		cmd := strings.Split(strings.Trim(message, "\n"), " ")
		switch cmd[0] {
		case "quit":
			break L
		case "start":
			s.ConnectoToServer("5.45.85.226:6584")
		case "search":
			s.Search(cmd[1]) // do not check len
		case "stop":
			s.DisconnectFromServer()
		case "slist":
			s.GetServerList()
		case "rep":
			//fmt.Println("Server connection status", sc.Status())
		default:
			fmt.Print("RESP", message)
		}
	}

	s.Stop()
}
