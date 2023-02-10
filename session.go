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
	configuration    Config
	comm             chan string
	server_packets   chan proto.Serializable
	wg               sync.WaitGroup
	listener         net.Listener
	connectionsMutex sync.Mutex
	connections      map[proto.Endpoint]*PeerConnection
	serverConnection ServerConnection
	ClientId         uint32
}

func CreateSession(config Config) *Session {
	serverPackets := make(chan proto.Serializable)
	return &Session{
		configuration:    config,
		comm:             make(chan string),
		connections:      make(map[proto.Endpoint]*PeerConnection),
		server_packets:   serverPackets,
		serverConnection: CreateServerConnection(serverPackets)}
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
		go s.accept(&s.listener)
	}

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
		case c, ok := <-s.server_packets:
			if ok {
				switch data := c.(type) {
				case *proto.SearchResult:
					fmt.Printf("session received search result size %d\n", data.Size())
					for _, x := range data.Items {
						a := proto.ToSearchItem(&x)
						fmt.Println("File", a.Filename, "size", a.Filesize, "sources", a.Sources, "complete sources", a.CompleteSources)
					}
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

	//for k, _ := range s.connections {
	//	k.conn.Close()
	//}

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

func (s *Session) accept(listener *net.Listener) {
	fmt.Println("Session listener started")
	for {
		_, e := (*listener).Accept()
		if e != nil {
			fmt.Printf("Accepting error %v\n", e)
			break
		} else {
			s.connectionsMutex.Lock()
			defer s.connectionsMutex.Unlock()
			//register_connection <- &SessionConnection{conn: c, err: e}
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

func (s *Session) GetServerList() {
	go s.serverConnection.ServerList()
}

func (s *Session) CreateHelloAnswer() proto.HelloAnswer {
	hello := proto.HelloAnswer{}
	hello.H = s.configuration.UserAgent
	hello.Point.Ip = s.ClientId
	hello.Point.Port = s.configuration.ListenPort

	hello.Properties = append(hello.Properties, proto.CreateTag(s.configuration.ClientName, proto.CT_NAME, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(s.configuration.ModName, proto.CT_MOD_VERSION, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(s.configuration.AppVersion, proto.CT_VERSION, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(0, proto.CT_EMULE_UDPPORTS, ""))
	// do not send CT_EM_VERSION since it will activate secure identification we are not support

	mo := proto.MiscOptions{}
	mo.UnicodeSupport = 1
	mo.DataCompVer = 0        // support data compression
	mo.NoViewSharedFiles = 1  // temp value
	mo.SourceExchange1Ver = 0 // SOURCE_EXCHG_LEVEL - important value

	mo2 := proto.MiscOptions2(0)
	mo2.SetCaptcha()
	mo2.SetLargeFiles()
	mo2.SetSourceExt2()
	version := makeFullED2KVersion(uint32(proto.SO_AMULE), s.configuration.ModMajorVersion, s.configuration.ModMinorVersion, s.configuration.ModBuildVersion)

	hello.Properties = append(hello.Properties, proto.CreateTag(version, proto.CT_EMULE_VERSION, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(mo.AsUint32(), proto.CT_EMULE_MISCOPTIONS1, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(mo2, proto.CT_EMULE_MISCOPTIONS2, ""))
	return hello
}

func makeFullED2KVersion(clientId uint32, a uint32, b uint32, c uint32) uint32 {
	return (clientId << 24) | (a << 17) | (b << 10) | (c << 7)
}

func (s *Session) ConnectoToPeer(endpoint proto.Endpoint) *PeerConnection {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	_, ok := s.connections[endpoint]
	if !ok {
		pc := PeerConnection{endpoint: endpoint}
		s.connections[endpoint] = &pc
		return &pc
	}

	return nil
}

func (s *Session) ClosePeerConnection(endpoint proto.Endpoint) {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	delete(s.connections, endpoint)
}
