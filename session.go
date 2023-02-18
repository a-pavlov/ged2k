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
	configuration   Config
	comm            chan string
	wg              sync.WaitGroup
	listener        net.Listener
	peerConnections []*PeerConnection
	transfers       []*Transfer

	// server section
	serverConnection           *ServerConnection
	serverPackets              chan proto.Serializable
	registerServerConnection   chan *ServerConnection
	unregisterServerConnection chan *ServerConnection

	// peer connection
	registerPeerConnection   chan *PeerConnection
	unregisterPeerConnection chan *PeerConnection

	ClientId uint32
	stat     Statistics
}

func CreateSession(config Config) *Session {
	return &Session{
		configuration:              config,
		comm:                       make(chan string),
		peerConnections:            make([]*PeerConnection, 0),
		serverPackets:              make(chan proto.Serializable),
		registerServerConnection:   make(chan *ServerConnection),
		unregisterServerConnection: make(chan *ServerConnection),
		registerPeerConnection:     make(chan *PeerConnection),
		unregisterPeerConnection:   make(chan *PeerConnection)}
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
	serverLastReq := time.Time{}
	serverLastRes := time.Time{}
	//serverCloseRequested := false

	lastTick := time.Time{}

	for execute {
		select {
		case sc := <-s.unregisterServerConnection:
			{
				fmt.Printf("Server connection closed, reason: \"%v\"\n", sc.lastError)
				s.serverConnection = nil
				serverConnected = false
				serverLastRes = time.Time{}
				serverLastReq = time.Time{}
				//serverCloseRequested = false

				if candidate != nil {
					go candidate.Start(s)
					candidate = nil
				}
			}
		case sc := <-s.registerServerConnection:
			{
				fmt.Println("Server connection established")
				s.serverConnection = sc
				if candidate != nil || serverRequestedDisconnect {
					fmt.Println("Server disconnect was requested")
					//serverCloseRequested = true
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
						go CreateServerConnection(elems[1]).Start(s)
					} else {
						candidate = CreateServerConnection(elems[1])
						if s.serverConnection != nil && serverConnected {
							//serverCloseRequested = true
							s.serverConnection.connection.Close()
						}
					}
				case "disconnect":
					if s.serverConnection != nil {
						if serverConnected {
							//serverCloseRequested = true
							s.serverConnection.connection.Close()
						} else {
							serverRequestedDisconnect = true
						}
					}
				case "search":
					if s.serverConnection != nil && serverConnected {
						parsed, err := proto.BuildEntries(0, 0, 0, 0, "", "", "", 0, 0, elems[1])
						if err == nil {
							req, err := proto.PackRequest(parsed)
							if err == nil {
								_, err = s.serverConnection.SendPacket(&req)
								if err == nil {
									serverLastReq = time.Now()
								}
							}
						}
					}
				case "serverlist":
					if s.serverConnection != nil && serverConnected && serverLastReq.IsZero() {
						req := proto.GetServerList{}
						_, err := s.serverConnection.SendPacket(&req)
						if err != nil {
							s.serverConnection.connection.Close()
						} else {
							serverLastReq = time.Now()
						}
					}
				case "peer":
					pc := PeerConnection{address: elems[1]}
					s.peerConnections = append(s.peerConnections, &pc)
					go pc.Start(s)
				default:
					fmt.Printf("Unknown command %s\n", cmd)
				}
			}
		case c, ok := <-s.serverPackets:
			if ok {
				serverLastRes = time.Now()
				serverLastReq = time.Time{}

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
					case *proto.ByteContainer:
						fmt.Println("Message from server", string(*data))
					default:
						fmt.Println("session: unknown server packet received")
					}
				}
			}
		case <-tick:
			fmt.Println("Tick")
			currentTime := time.Now()
			if s.serverConnection != nil {
				if !serverLastReq.IsZero() && currentTime.Sub(serverLastReq).Seconds() > 5 {
					serverLastReq = time.Time{}
				}

				if !serverLastRes.IsZero() && currentTime.Sub(serverLastRes).Seconds() > 40 {
					// no response for long time
				}

				// send statistics here
			}

			// enumerate transfer to get new peers
			stepsSinceLastConnect := 0
			connectionsReserve := s.configuration.MaxConnectsPerSecond
			enumerateCandidates := true
			if len(s.transfers) > 0 && len(s.peerConnections) < s.configuration.MaxConnections {
				for enumerateCandidates {
					for _, t := range s.transfers {
						if t.WantMorePeers() {
							candidate := t.policy.FindConnectCandidate(currentTime)
							if candidate != nil {
								peerConnection := s.GetPeerConnectionByEndpoint(candidate.endpoint)
								if peerConnection == nil {
									candidate.LastConnected = currentTime
									candidate.NextConnection = time.Time{}
									peerConnection := PeerConnection{address: candidate.endpoint.AsString(), transfer: t, peer: candidate}
									s.peerConnections = append(s.peerConnections, &peerConnection)
									t.connections = append(t.connections, &peerConnection)
									candidate.peerConnection = &peerConnection
									connectionsReserve--
									stepsSinceLastConnect = 0
									go peerConnection.Start(s)
								} else {
									// peer connection on this endpoint already connected
									// update next time
								}
							}
							// transfer policy find connect candidate
							// if connected

						}
					}

					stepsSinceLastConnect++

					// if we have gone two whole loops without
					// handing out a single connection, break
					if stepsSinceLastConnect > len(s.transfers)*2 {
						enumerateCandidates = false
						break
					}

					// if we should not make any more connections
					if connectionsReserve == 0 {
						enumerateCandidates = false
						break
					}
				}
			}

			// tick to collect statistics
			for _, x := range s.transfers {
				if !lastTick.IsZero() {
					x.SecondTick(currentTime.Sub(lastTick), s)
				}
			}

			lastTick = currentTime

		case peerConnection := <-s.registerPeerConnection:
			s.peerConnections = append(s.peerConnections, peerConnection)
			if peerConnection.transfer == nil {
				//looking for corresponding transfer
				// policy - newConnection
				//peerConnection.transfer.
			}
		case peerConnection := <-s.unregisterPeerConnection:
			s.peerConnections = removePeerConnection(peerConnection, s.peerConnections)
			if peerConnection.transfer != nil {
				peerConnection.transfer.connections = removePeerConnection(peerConnection, peerConnection.transfer.connections)
			}

			if peerConnection.peer != nil {
				peerConnection.peer.peerConnection = nil
				peerConnection.peer.LastConnected = time.Now()
				// check error somehow
				if peerConnection.lastError != nil {
					peerConnection.peer.FailCount += 1
				}
			}

			peerConnection.transfer = nil
			peerConnection.peer = nil
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
		conn, e := (*listener).Accept()
		if e != nil {
			fmt.Printf("Accepting error %v\n", e)
			break
		} else {
			pc := PeerConnection{connection: conn}
			go pc.Start(s)
			s.registerPeerConnection <- &pc
		}
	}
}

func (s *Session) Search(keyword string) {
	s.comm <- "search " + keyword
}

func (s *Session) GetServerList() {
	s.comm <- "serverlist"
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

func (s *Session) Connect(address string) {
	s.comm <- "connect " + address
}

func (s *Session) Disconnect() {
	s.comm <- "disconnect"
}

func (s *Session) DiscardPacket() {
	s.serverPackets <- nil
}

func (s *Session) Send(serverConnection *ServerConnection, data proto.Serializable) {
	_, serverConnection.lastError = serverConnection.SendPacket(data)
	if serverConnection.lastError != nil {
		serverConnection.connection.Close()
	}
}

func (s *Session) GetPeerConnectionByEndpoint(endpoint proto.Endpoint) *PeerConnection {
	for _, x := range s.peerConnections {
		if x.endpoint == endpoint {
			return x
		}
	}

	return nil
}
