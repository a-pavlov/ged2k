package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

type Session struct {
	configuration    Config
	comm             chan string
	wg               sync.WaitGroup
	listener         net.Listener
	connectionsMutex sync.Mutex
	connections      map[proto.Endpoint]*PeerConnection

	// server section
	serverConnection           *ServerConnection
	serverPackets              chan proto.Serializable
	registerServerConnection   chan *ServerConnection
	unregisterServerConnection chan *ServerConnection

	ClientId uint32
}

func CreateSession(config Config) *Session {
	return &Session{
		configuration:              config,
		comm:                       make(chan string),
		connections:                make(map[proto.Endpoint]*PeerConnection),
		serverPackets:              make(chan proto.Serializable),
		registerServerConnection:   make(chan *ServerConnection),
		unregisterServerConnection: make(chan *ServerConnection)}
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

	serverConnected := false
	serverRequestedDisconnect := false
	var candidate *ServerConnection

	for execute {
		select {
		case <-s.unregisterServerConnection:
			{
				fmt.Println("Server connection closed")
				s.serverConnection = nil
				serverConnected = false
				if candidate != nil {
					go s.ConnectToServer(candidate)
					candidate = nil
				}
			}
		case sc := <-s.registerServerConnection:
			{
				fmt.Println("Server connection established")
				s.serverConnection = sc
				if candidate != nil || serverRequestedDisconnect {
					fmt.Println("Server disconnect was requested")
					s.serverConnection.connection.Close()
				}
				serverConnected = true
				serverRequestedDisconnect = false
			}
		case cmd, ok := <-s.comm:
			if !ok {
				fmt.Println("Session exit requested")
				execute = false
			} else {
				elems := strings.Split(cmd, " ")

				switch elems[0] {
				case "hello":
					fmt.Println("Hello !!!")
				case "connect":
					fmt.Println("Requested connect to", elems[1])
					if s.serverConnection == nil {
						go s.ConnectToServer(CreateServerConnection(elems[1]))
					} else {
						candidate = CreateServerConnection(elems[1])
						if s.serverConnection != nil && serverConnected {
							s.serverConnection.connection.Close()
						}
					}
				case "disconnect":
					if s.serverConnection != nil {
						if serverConnected {
							s.serverConnection.connection.Close()
						} else {
							serverRequestedDisconnect = true
						}
					}
				case "search":
					if s.serverConnection != nil && serverConnected {
						s.serverConnection.Search(elems[1])
					}
				default:
					fmt.Printf("Unknown command %s\n", cmd)
				}
			}
		case c, ok := <-s.serverPackets:
			if ok {
				//req := s.serverConnection.outgoingOrder[0]
				if len(s.serverConnection.outgoingOrder) > 0 {
					s.serverConnection.outgoingOrder = s.serverConnection.outgoingOrder[1:]
				}
				if c != nil {
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
			}
		case <-tick:
			fmt.Println("Tick")
			currentTime := time.Now()
			if s.serverConnection != nil {
				if !s.serverConnection.lastSend.IsZero() && currentTime.Sub(s.serverConnection.lastSend).Seconds() > 10 {
					fmt.Printf("ServerConnection Tick. Outgoing order size: %d", len(s.serverConnection.outgoingOrder))
					s.serverConnection.lastSend = time.Time{}
					s.serverPackets <- nil
				}
			}
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
	if s.serverConnection != nil {
		s.serverConnection.connection.Close()
	}
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

func (s *Session) Search(keyword string) {
	s.comm <- "search " + keyword
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

func (s *Session) ConnectToServer(serverConnection *ServerConnection) {
	connection, err := net.Dial("tcp", serverConnection.address)
	if err != nil {
		fmt.Println("Connect error", err)
		s.unregisterServerConnection <- serverConnection
		return
	}

	fmt.Println("Connected!")
	serverConnection.connection = connection

	s.registerServerConnection <- serverConnection

	var version uint32 = 0x3c
	var versionClient uint32 = (proto.GED2K_VERSION_MAJOR << 24) | (proto.GED2K_VERSION_MINOR << 17) | (proto.GED2K_VERSION_TINY << 10) | (1 << 7)
	var capability uint32 = proto.CAPABLE_AUXPORT | proto.CAPABLE_NEWTAGS | proto.CAPABLE_UNICODE | proto.CAPABLE_LARGEFILES | proto.CAPABLE_ZLIB

	var hello proto.UsualPacket
	hello.H = proto.EMULE
	hello.Point = proto.Endpoint{Ip: 0, Port: 20033}
	hello.Properties = append(hello.Properties, proto.CreateTag(version, proto.CT_VERSION, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(capability, proto.CT_SERVER_FLAGS, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag("ged2k", proto.CT_NAME, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(versionClient, proto.CT_EMULE_VERSION, ""))

	n, err := serverConnection.SendPacket(&hello)

	if err != nil {
		fmt.Printf("Error write to socket %v\n", err)
		s.unregisterServerConnection <- serverConnection
		return
	} else {
		fmt.Printf("Bytes %d have been written\n", n)
	}

	pc := proto.PacketCombiner{}
	for {
		ph, bytes, err := pc.Read(connection)
		if err != nil {
			fmt.Printf("Can not read bytes from server %v", err)
			break
		}

		sb := proto.StateBuffer{Data: bytes}

		switch ph.Packet {
		case proto.OP_SERVERLIST:
			elems := sb.ReadUint8()
			if sb.Error() == nil && elems < 100 {
				c := proto.Collection{}
				for i := 0; i < int(elems); i++ {
					c = append(c, &proto.Endpoint{})
				}
				sb.Read(&c)
			}
		case proto.OP_GETSERVERLIST:
			// ignore
		case proto.OP_SERVERMESSAGE:
			bc := proto.ByteContainer{}
			bc.Get(&sb)
			if sb.Error() == nil {
				s.serverPackets <- &bc
				fmt.Println("Receive message from server", string(bc))
			}
		case proto.OP_SERVERSTATUS:
			ss := proto.Status{}
			ss.Get(&sb)
			if sb.Error() == nil {
				s.serverPackets <- &ss
				fmt.Println("Server status files:", ss.FilesCount, "users", ss.UsersCount)
			}
		case proto.OP_IDCHANGE:
			idc := proto.IdChange{}
			idc.Get(&sb)
			if sb.Error() == nil {
				fmt.Println("Server id change", idc.ClientId)
				s.serverPackets <- &idc
			}
		case proto.OP_SERVERIDENT:
			p := proto.UsualPacket{}
			p.Get(&sb)
			if sb.Error() == nil {
				fmt.Println("Received server info packet")
			}
		case proto.OP_SEARCHRESULT:
			p := proto.SearchResult{}
			p.Get(&sb)
			if sb.Error() == nil {
				fmt.Printf("Search result received: %d, more results %v\n", len(p.Items), p.MoreResults)
				s.serverPackets <- &p
			} else {
				fmt.Printf("Unable to de-serealize %v\n", sb.Error())
			}
		case proto.OP_SEARCHREQUEST:
			// ignore
		case proto.OP_QUERY_MORE_RESULT:
			// ignore - out only
		case proto.OP_GETSOURCES:
			// ignore - out only
		case proto.OP_FOUNDSOURCES:
			// ignore
		case proto.OP_CALLBACKREQUEST:
			// ignore - out
		case proto.OP_CALLBACKREQUESTED:
			// ignore
		case proto.OP_CALLBACK_FAIL:
			// ignore
		default:
			fmt.Printf("Packet %x", bytes)
		}

		if sb.Error() != nil {
			fmt.Printf("Error on packet read %v", sb.Error())
			break
		}

		// finalize server connection status
		serverConnection.lastSend = time.Time{}
		serverConnection.lastReceived = time.Now()
		if len(serverConnection.outgoingOrder) > 0 {
			serverConnection.outgoingOrder = serverConnection.outgoingOrder[1:]
		}

		fmt.Printf("Data received. Outgoing order size %d\n", len(serverConnection.outgoingOrder))
	}

	fmt.Println("Exit server connection procedure")
	s.unregisterServerConnection <- serverConnection
}

func (s *Session) Connect(address string) {
	s.comm <- "connect " + address
}

func (s *Session) Disconnect() {
	s.comm <- "disconnect"
}
