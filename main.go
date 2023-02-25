package main

import (
	"bufio"
	"fmt"
	"github.com/a-pavlov/ged2k/proto"
	"os"
	"strings"
)

func main() {
	fmt.Println("Hello ged2k")
	reader := bufio.NewReader(os.Stdin)
	cfg := Config{UserAgent: proto.EMULE, ListenPort: 4888, Name: "TestGed2k", MaxConnections: 100, ModName: "jed2k", ClientName: "jed2k", AppVersion: 0x3c}
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

			//s.Connect("176.123.5.89:4725")
			s.Connect("5.45.85.226:6584")
		case "search":
			s.Search(cmd[1]) // do not check len
		case "stop":
			s.Disconnect()
		case "slist":
			s.GetServerList()
		case "rep":
			//fmt.Println("Server connection status", sc.Status())
		default:
			s.Cmd(strings.Trim(message, "\n"))
		}
	}

	s.Stop()
}

// tran /tmp/test.txt 460359517F89AE010793896EDE7D30F8 4
// OP_PUBLICIP_REQ
