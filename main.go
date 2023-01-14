package main

import (
	"bufio"
	"fmt"
	"os"
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
		switch message {
		case "quit\n":
			break L
		case "start\n":
			go s.ConnectoToServer("5.45.85.226:6584")
		case "search\n":
			go s.Search("game")
		case "stop\n":
			s.DisconnectFromServer()
		case "rep\n":
			//fmt.Println("Server connection status", sc.Status())
		default:
			fmt.Print("RESP", message)
		}
	}

	s.Stop()
}
