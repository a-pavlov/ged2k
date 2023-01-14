package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

type SessionConnection struct {
	config Config
	conn   net.Conn
	err    error
}

type Session struct {
	configuration         Config
	comm                  chan string
	register_connection   chan *SessionConnection
	unregister_connection chan *SessionConnection
	server_packets        chan proto.Serializable
	wg                    sync.WaitGroup
	listener              net.Listener
	connections           map[*SessionConnection]bool
	serverConnection      ServerConnection
}

func CreateSession(config Config) *Session {
	serverPackets := make(chan proto.Serializable)
	return &Session{
		configuration:         config,
		comm:                  make(chan string),
		register_connection:   make(chan *SessionConnection),
		unregister_connection: make(chan *SessionConnection),
		connections:           make(map[*SessionConnection]bool),
		server_packets:        serverPackets,
		serverConnection:      CreateServerConnection(serverPackets)}
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
		case c, ok := <-s.server_packets:
			if ok {
				switch data := c.(type) {
				case *proto.SearchResult:
					fmt.Printf("session received search result size %d\n", data.Size())
				case *proto.FoundFileSources:
					fmt.Printf("session found file sources %d\n", data.Size())
				default:
					fmt.Println("session: unknown server packet received")
				}
			}
		case <-tick:
			fmt.Println("Tick")
			currentTime := time.Now()
			s.serverConnection.Tick(currentTime)
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
	fmt.Println("Session stop requested")
	close(s.comm)
	s.serverConnection.Stop()
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

func (s *Session) ConnectoToServer(address string) {
	if s.serverConnection.IsConnected() {
		s.serverConnection.Stop()
	}

	go s.serverConnection.Start(address)
}

func (s *Session) DisconnectFromServer() {
	go s.serverConnection.Stop()
}

func (s *Session) Search(keyword string) {
	go s.serverConnection.Search(keyword)
}
