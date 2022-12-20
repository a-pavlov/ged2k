package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type SessionConnection struct {
	config Config
	conn   net.Conn
	err    error
}

type Session struct {
	comm                  chan string
	register_connection   chan *SessionConnection
	unregister_connection chan *SessionConnection
	wg                    sync.WaitGroup
	listener              net.Listener
	connections           map[*SessionConnection]bool
}

func (s *Session) Tick() {
	tick := time.Tick(5000 * time.Millisecond)
	execute := true
	s.wg.Add(1)
	defer s.wg.Done()

	// start listener
	var e error
	s.listener, e = net.Listen("tcp", ":12345")
	if e != nil {
		// can not listen

	} else {
		go s.accept(&s.listener, s.register_connection)
	}

E:
	for execute {
		select {
		case cmd, ok := <-s.comm:
			if !ok {
				fmt.Println("Session exit requested")
				execute = false
			} else {
				switch cmd {
				case "hello":
					fmt.Println("Hello !!!")
				default:
					fmt.Printf("Unknown command %s\n", cmd)
				}
			}
		case c, ok := <-s.register_connection:
			if ok {
				fmt.Printf("Incoming connection %v\n", c)
				s.connections[c] = true
				go s.receive(c)
			} else {
				// need channel replace or null
				fmt.Println("Listener failed, need channel replace")
				break E
			}
		case conn, ok := <-s.unregister_connection:
			if ok {
				if s.connections[conn] {
					fmt.Println("Connection closed")
					delete(s.connections, conn)
				}
			}
		case <-tick:
			fmt.Println("Tick")
		}
	}

	e = s.listener.Close()
	if e != nil {
		fmt.Printf("Listener stop error %v\n", e)
	}

	for k, _ := range s.connections {
		k.conn.Close()
	}

	fmt.Println("Session closed")
}

func (s *Session) Start() {
	go s.Tick()
}

func (s *Session) Stop() {
	close(s.comm)
	s.wg.Wait()
}

func (s *Session) receive(sc *SessionConnection) {
	for {
		message := make([]byte, 4096)
		length, err := sc.conn.Read(message)
		if err != nil {
			s.unregister_connection <- sc
			sc.conn.Close()
			break
		}

		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func (s *Session) accept(listener *net.Listener, register_connection chan *SessionConnection) {

	fmt.Println("Session listener started")
	for {
		c, e := (*listener).Accept()
		if e != nil {
			fmt.Printf("Accepting error %v\n", e)
			close(register_connection)
			break
		} else {
			register_connection <- &SessionConnection{conn: c, err: e}
		}
	}
}

func (s *Session) connect(address string, channel chan interface{}) {
	//connection, err := net.Dial("tcp", address)
	//if err != nil {
	//}
}
