package main

import (
	"fmt"
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
	peerConnections map[proto.Endpoint]*PeerConnection
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
		peerConnections:            make(map[proto.Endpoint]*PeerConnection, 0),
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

	var candidate *ServerConnection

	lastTick := time.Time{}

	for execute {
		select {
		case sc := <-s.unregisterServerConnection:
			{
				log.Printf("Server connection closed, reason: \"%v\"\n", sc.lastError)
				s.serverConnection = nil
				if candidate != nil {
					s.serverConnection = candidate
					go candidate.Start(s)
					candidate = nil
				}
			}
		case sc := <-s.registerServerConnection:
			{
				log.Println("Server connection established")
				sc.Connected = true
				sc.LastReceivedTime = time.Now().Add(time.Duration(30) * time.Second)

				if candidate != nil || sc.DisconnectRequested {
					log.Println("Server disconnect was requested")
					//serverCloseRequested = true
					s.serverConnection.connection.Close()
				}
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
						s.serverConnection = NewServerConnection(elems[1])
						go s.serverConnection.Start(s)
					} else {
						candidate = NewServerConnection(elems[1])
						if s.serverConnection != nil && s.serverConnection.Connected {
							if !s.serverConnection.DisconnectRequested {
								go s.serverConnection.Close()
							}
							s.serverConnection.DisconnectRequested = true
						}
					}
				case "disconnect":
					if s.serverConnection != nil {
						if s.serverConnection.Connected && !s.serverConnection.DisconnectRequested {
							go s.serverConnection.Close()
						}
						s.serverConnection.DisconnectRequested = true
					}

					for ep, x := range s.peerConnections {
						fmt.Printf("REQ ds %s\n", ep.AsString())
						x.Close(true)
					}
				case "search":
					if s.serverConnection != nil && s.serverConnection.Connected {
						parsed, err := proto.BuildEntries(0, 0, 0, 0, "", "", "", 0, 0, elems[1])
						if err == nil {
							req, err := proto.PackRequest(parsed)
							if err == nil {
								_, err = s.serverConnection.SendPacket(&req)
								//if err == nil {
								//	serverLastReq = time.Now()
								//}
							}
						}
					}
				case "serverlist":
					if s.serverConnection != nil && s.serverConnection.Connected {
						req := proto.GetServerList{}
						s.serverConnection.SendPacket(&req)
						/*if err != nil {
							s.serverConnection.connection.Close()
						} else {
							serverLastReq = time.Now()
						}*/
					}
				case "peer":
					pc := PeerConnection{Address: elems[1]}
					ep, _ := proto.FromString(elems[1])
					s.peerConnections[ep] = &pc
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
						if tran.policy.AddPeer(&Peer{endpoint: proto.EndpointFromString("127.0.0.1:4662"), SourceFlag: 'S'}) {
							log.Printf("Added peer 127.0.0.1:4662\n")
						}
						if tran.policy.AddPeer(&Peer{endpoint: proto.EndpointFromString("127.0.0.1:4663"), SourceFlag: 'S'}) {
							log.Printf("Added peer 127.0.0.1:4663\n")
						}

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
				if s.serverConnection != nil {
					s.serverConnection.LastReceivedTime = time.Now()
				}

				log.Printf("server last recv time %v\n", s.serverConnection.LastReceivedTime)

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
					case *proto.Status:
						log.Printf("Server status[users: %d, files:%d]\n", data.UsersCount, data.FilesCount)
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
			currentTime := time.Now()
			if s.serverConnection != nil {

				if !s.serverConnection.LastReceivedTime.IsZero() &&
					currentTime.After(s.serverConnection.LastReceivedTime.Add(time.Duration(30)*time.Second)) &&
					currentTime.After(s.serverConnection.LastSendTime.Add(time.Duration(30)*time.Second)) {
					log.Printf("server connection ping required, last recv time %v last send time %v\n",
						s.serverConnection.LastReceivedTime, s.serverConnection.LastSendTime)
					sl := proto.GetServerList{}
					_, err := s.serverConnection.SendPacket(&sl)
					if err != nil {
						log.Printf("server packet send error: %v\n", err)
					}
				}

				if s.configuration.ServerReconnectTimeoutSec > 0 &&
					!s.serverConnection.LastReceivedTime.IsZero() &&
					currentTime.After(s.serverConnection.LastReceivedTime.Add(time.Duration(s.configuration.ServerReconnectTimeoutSec)*time.Second)) {
					log.Printf("server connection no answer for a long time %v last send time %v - reconnect required",
						s.serverConnection.LastReceivedTime, s.serverConnection.LastSendTime)
					// no answer from server connection for a long time, reconnect
					candidate = NewServerConnection(s.serverConnection.address)
					s.serverConnection.DisconnectRequested = true
					go s.serverConnection.Close()
				}
			}

			// enumerate transfer to get new peers
			stepsSinceLastConnect := 0
			connectionsReserve := s.configuration.MaxConnectsPerSecond
			enumerateCandidates := true
			if len(s.transfers) > 0 && len(s.peerConnections) < s.configuration.MaxConnections {
				for enumerateCandidates {
					for _, transfer := range s.transfers {
						if transfer.WantMoreSources(currentTime) {
							if s.serverConnection != nil && s.serverConnection.Connected {
								req := proto.GetFileSources{Hash: transfer.Hash}
								go s.serverConnection.SendPacket(&req)
								// request next time in one minute
								transfer.RequestSourcesNextTime = time.Now().Add(time.Minute * time.Duration(1))
							}
						}
						if transfer.WantMorePeers() {
							candidate := transfer.policy.FindConnectCandidate(currentTime)
							if candidate != nil {
								_, ok := s.peerConnections[candidate.endpoint]
								if !ok {
									candidate.LastConnected = currentTime
									peerConnection := NewPeerConnection(candidate.endpoint.AsString(), transfer, candidate)
									s.peerConnections[candidate.endpoint] = peerConnection
									candidate.peerConnection = peerConnection
									connectionsReserve--
									stepsSinceLastConnect = 0
									go peerConnection.Start(s)
								} else {
									// move next connection on peer to the future to avoid returning it as candidate
									fmt.Printf("candidate already in session: %s", candidate.endpoint)
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
			fmt.Printf("register peer connection %s", peerConnection.Address)
			s.peerConnections[peerConnection.endpoint] = peerConnection
			if peerConnection.transfer == nil {
				//looking for corresponding transfer
				// policy - newConnection
				//peerConnection.transfer.
			}
		case peerConnectionPacket := <-s.unregisterPeerConnection:
			log.Printf("unregister peer connection, peer %v", peerConnectionPacket.Connection.peer)
			delete(s.peerConnections, peerConnectionPacket.Connection.endpoint)

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
			for _, x := range s.peerConnections {
				if x.transfer == transfer {
					go x.Close(true)
				}
			}
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

	if s.serverConnection != nil {
		go s.serverConnection.Close()
	}

	close(s.comm)
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
	mo.DataCompVer = 1        // support data compression
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

func (s *Session) GetPeerConnectionByEndpoint(endpoint proto.Endpoint) *PeerConnection {
	for _, x := range s.peerConnections {
		if x.endpoint == endpoint {
			return x
		}
	}

	return nil
}
