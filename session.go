package main

import (
	"io/ioutil"
	"log"
	"net"
	"strconv"
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
	transfers       map[proto.ED2KHash]*Transfer

	// server section
	serverConnection           *ServerConnection
	serverPackets              chan proto.Serializable
	registerServerConnection   chan *ServerConnection
	unregisterServerConnection chan *ServerConnection

	// peer connection
	registerPeerConnection   chan *PeerConnection
	unregisterPeerConnection chan PeerConnectionPacket

	//transfer
	transferChanResumeDataRead chan *Transfer
	transferChanFinished       chan *Transfer
	transferChanPaused         chan *Transfer
	transferResumeData         chan proto.AddTransferParameters
	transferChanError          chan TransferError

	statistics      Statistics
	statReceiveChan chan StatPacket
	statSendChan    chan StatPacket

	ClientId uint32
	Stat     Statistics
}

func NewSession(config Config) *Session {
	log.Println("Create session")
	return &Session{
		configuration:              config,
		comm:                       make(chan string),
		peerConnections:            make([]*PeerConnection, 0),
		serverPackets:              make(chan proto.Serializable),
		registerServerConnection:   make(chan *ServerConnection),
		unregisterServerConnection: make(chan *ServerConnection),
		registerPeerConnection:     make(chan *PeerConnection),
		unregisterPeerConnection:   make(chan PeerConnectionPacket),
		transfers:                  make(map[proto.ED2KHash]*Transfer),
		transferChanResumeDataRead: make(chan *Transfer),
		transferChanFinished:       make(chan *Transfer),
		transferChanPaused:         make(chan *Transfer),
		transferResumeData:         make(chan proto.AddTransferParameters),
		transferChanError:          make(chan TransferError),
		statReceiveChan:            make(chan StatPacket),
		statSendChan:               make(chan StatPacket),
		Stat:                       MakeStatistics(),
	}
}

func (s *Session) Tick() {
	tick := time.Tick(1000 * time.Millisecond)
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
				log.Printf("Server connection closed, reason: \"%v\"\n", sc.lastError)
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
				log.Println("Server connection established")
				s.serverConnection = sc
				if candidate != nil || serverRequestedDisconnect {
					log.Println("Server disconnect was requested")
					//serverCloseRequested = true
					s.serverConnection.connection.Close()
				}
				serverConnected = true
				serverRequestedDisconnect = false
			}
		case cmd, ok := <-s.comm:
			if !ok {
				log.Println("Session exit requested")
				execute = false
			} else {
				log.Println("Session received cmd:", cmd)
				elems := strings.Split(cmd, " ")

				switch elems[0] {
				case "hello":
					log.Println("Hello !!!")
				case "connect":
					log.Println("Requested connect to", elems[1])
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
					pc := PeerConnection{Address: elems[1]}
					s.peerConnections = append(s.peerConnections, &pc)
					go pc.Start(s)
				case "load":
					for i := 1; i < len(elems); i++ {
						rd, err := ioutil.ReadFile(elems[1])
						if err != nil {
							log.Println("Error read resume data file", err)
						} else {
							sb := proto.StateBuffer{Data: rd}
							var atp proto.AddTransferParameters
							sb.Read(&atp)
							if sb.Error() == nil {
								t := NewTransfer(atp.Hashes.Hash, atp.Filename.ToString(), atp.Filesize)
								go t.Start(s, &atp)
							} else {
								log.Printf("Can not read resume data file %v\n", sb.Error())
							}
						}
					}
				case "tran":
					log.Println("tran command accepted")
					filename := elems[1]
					hash := proto.String2Hash(elems[2])
					size, err := strconv.ParseUint(elems[3], 10, 64)
					if err == nil {
						log.Printf(" add transfer %v to file %s\n", hash.ToString(), filename)
						tran := NewTransfer(hash, filename, size)
						s.transfers[hash] = tran
						tran.policy.AddPeer(&Peer{endpoint: proto.EndpointFromString("127.0.0.1:4662"), SourceFlag: 'S'})
						log.Println("Added peer to transfer")
						go tran.Start(s, nil)
					} else {
						log.Println("Error on transfer adding", err)
					}

				default:
					log.Printf("Unknown command %s\n", cmd)
				}
			}
		case c, ok := <-s.serverPackets:
			if ok {
				serverLastRes = time.Now()
				serverLastReq = time.Time{}

				if c != nil {
					switch data := c.(type) {
					case *proto.SearchResult:
						log.Printf("session received search result size %d\n", data.Size())
						for _, x := range data.Items {
							a := proto.ToSearchItem(&x)
							log.Println("File", a.Filename, "size", a.Filesize, "sources", a.Sources, "complete sources", a.CompleteSources)
						}
					case *proto.FoundFileSources:
						log.Printf("session found file sources %d\n", data.Size())
						transfer, ok := s.transfers[data.Hash]
						if ok {
							log.Printf("Got sources for %s\n", data.Hash)
							for _, x := range data.Sources {
								if transfer.policy.AddPeer(&Peer{SourceFlag: PEER_SRC_SERVER, endpoint: x}) {
									log.Printf("Transfer %s added source %s\n", data.Hash.ToString(), x.AsString())
								} else {
									log.Printf("Can not add peer %s to transfer %s\n", x.AsString(), data.Hash.ToString())
								}
							}
						} else {
							log.Printf("Got sources for %s, but can not find corresponding transfer\n", data.Hash.ToString())
						}
					case *proto.ByteContainer:
						log.Println("Message from server", string(*data))
					default:
						log.Println("session: unknown server packet received")
					}
				}
			}
		case statPacket := <-s.statReceiveChan:
			statPacket.Connection.Stat.ReceiveBytes(statPacket.Counter)
			s.Stat.ReceiveBytes(statPacket.Counter)
			if statPacket.Connection.transfer != nil {
				statPacket.Connection.transfer.Stat.ReceiveBytes(statPacket.Counter)
			}
		case statPacket := <-s.statSendChan:
			statPacket.Connection.Stat.SendBytes(statPacket.Counter)
			s.Stat.SendBytes(statPacket.Counter)
			if statPacket.Connection.transfer != nil {
				statPacket.Connection.transfer.Stat.SendBytes(statPacket.Counter)
			}
		case <-tick:
			//for i, x := range s.transfers {
			//	x.policy.AddPeer(&Peer{endpoint: proto.EndpointFromString("127.0.0.1:4662"), SourceFlag: 'S', Connectable: true})
			//	log.Println("Add peer to", i)
			//}
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
					for _, transfer := range s.transfers {
						if transfer.WantMoreSources(currentTime) {
							if s.serverConnection != nil && serverConnected {
								req := proto.GetFileSources{Hash: transfer.Hash}
								go s.serverConnection.SendPacket(&req)
								// request next time in one minute
								transfer.RequestSourcesNextTime = time.Now().Add(time.Minute * time.Duration(1))
							}
						}
						if transfer.WantMorePeers() {
							candidate := transfer.policy.FindConnectCandidate(currentTime)
							if candidate != nil {
								peerConnection := s.GetPeerConnectionByEndpoint(candidate.endpoint)
								if peerConnection == nil {
									candidate.LastConnected = currentTime
									peerConnection := NewPeerConnection(candidate.endpoint.AsString(), transfer, candidate)
									s.peerConnections = append(s.peerConnections, peerConnection)
									transfer.connections = append(transfer.connections, peerConnection)
									candidate.peerConnection = peerConnection
									connectionsReserve--
									stepsSinceLastConnect = 0
									go peerConnection.Start(s)
								} else {
									// move next connection on peer to the future to avoid returning it as candidate
									candidate.NextConnection = currentTime.Add(time.Second * time.Duration(15))
								}
							}
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
			dur := currentTime.Sub(lastTick)
			for _, x := range s.peerConnections {
				x.Stat.SecondTick(dur)
			}
			s.Stat.SecondTick(dur)
			for _, x := range s.transfers {
				x.Stat.SecondTick(dur)
			}

			lastTick = currentTime

		case peerConnection := <-s.registerPeerConnection:
			s.peerConnections = append(s.peerConnections, peerConnection)
			if peerConnection.transfer == nil {
				//looking for corresponding transfer
				// policy - newConnection
				//peerConnection.transfer.
			}
		case peerConnectionPacket := <-s.unregisterPeerConnection:
			s.peerConnections = removePeerConnection(peerConnectionPacket.Connection, s.peerConnections)
			if peerConnectionPacket.Connection.transfer != nil {
				peerConnectionPacket.Connection.transfer.connections = removePeerConnection(peerConnectionPacket.Connection, peerConnectionPacket.Connection.transfer.connections)
				// abort all blocks we have requested for now
				for _, x := range peerConnectionPacket.Connection.requestedBlocks {
					peerConnectionPacket.Connection.transfer.piecePicker.AbortBlock(x.block, peerConnectionPacket.Connection.peer)
				}
			}

			if peerConnectionPacket.Connection.peer != nil {
				peerConnectionPacket.Connection.peer.peerConnection = nil
				peerConnectionPacket.Connection.peer.LastConnected = time.Now()
				// check error somehow
				if !peerConnectionPacket.Connection.closedByRequest && peerConnectionPacket.Error != nil {
					peerConnectionPacket.Connection.peer.FailCount += 1
				}
			}

			peerConnectionPacket.Connection.transfer = nil
			peerConnectionPacket.Connection.peer = nil
		case transfer := <-s.transferChanFinished:
			transfer.Finished = true
			for _, x := range transfer.connections {
				go x.Close(true)
			}
			transfer.connections = transfer.connections[:0]
		case transfer := <-s.transferChanPaused:
			transfer.Paused = true
		// close all transfer's peers
		case transfer := <-s.transferChanResumeDataRead:
			transfer.ReadingResumeData = false
		case atp := <-s.transferResumeData:
			log.Println("Save resume data - ignore for now", atp.Filename.ToString())
		case te := <-s.transferChanError:
			te.transfer.LastError = te.err
			log.Printf("Transfer %s error %v\n", te.transfer.Hash.ToString(), te.err)
		}
	}

	e = s.listener.Close()
	if e != nil {
		log.Printf("Listener stop error %v\n", e)
	}

	//for k, _ := range s.connections {
	//	k.conn.Close()
	//}

	log.Println("Session closed")
}

func (s *Session) Start() {
	go s.Tick()
}

func (s *Session) Stop() {
	log.Println("Session stop requested")
	for _, x := range s.transfers {
		x.Stop()
	}

	// close all peer connections
	for _, x := range s.peerConnections {
		x.Close(true)
	}

	close(s.comm)
	if s.serverConnection != nil {
		s.serverConnection.connection.Close()
	}
	s.wg.Wait()
}

func (s *Session) accept(listener *net.Listener) {
	log.Println("Session listener started")
	for {
		conn, e := (*listener).Accept()
		if e != nil {
			log.Printf("Accepting error %v\n", e)
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
	hello.Hash = s.configuration.UserAgent
	hello.Point.Ip = s.ClientId
	hello.Point.Port = s.configuration.ListenPort

	hello.Properties = append(hello.Properties, proto.CreateTag(s.configuration.ClientName, proto.CT_NAME, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(s.configuration.ModName, proto.CT_MOD_VERSION, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(s.configuration.AppVersion, proto.CT_VERSION, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(uint32(0), proto.CT_EMULE_UDPPORTS, ""))
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
	hello.Properties = append(hello.Properties, proto.CreateTag(uint32(mo2), proto.CT_EMULE_MISCOPTIONS2, ""))
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

func (s *Session) Cmd(cmd string) {
	s.comm <- cmd
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
