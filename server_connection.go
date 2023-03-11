package main

import (
	"log"
	"net"
	"reflect"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

type ServerConnection struct {
	buffer     []byte
	connection net.Conn
	address    string
	lastError  error

	Connected           bool
	DisconnectRequested bool
	LastReceivedTime    time.Time
	LastSendTime        time.Time
}

func NewServerConnection(a string) *ServerConnection {
	return &ServerConnection{buffer: make([]byte, 200), address: a}
}

func (serverConnection *ServerConnection) Start(s *Session) {
	log.Println("Server conn init", time.Now())
	connection, err := net.Dial("tcp", serverConnection.address)
	if err != nil {
		serverConnection.lastError = err
		s.unregisterServerConnection <- serverConnection
		return
	}

	log.Println("Connected!", time.Now())
	serverConnection.connection = connection

	s.registerServerConnection <- serverConnection

	var version uint32 = 0x3c
	var versionClient uint32 = (proto.GED2K_VERSION_MAJOR << 24) | (proto.GED2K_VERSION_MINOR << 17) | (proto.GED2K_VERSION_TINY << 10) | (1 << 7)
	var capability uint32 = proto.CAPABLE_AUXPORT | proto.CAPABLE_NEWTAGS | proto.CAPABLE_UNICODE | proto.CAPABLE_LARGEFILES | proto.CAPABLE_ZLIB

	var hello proto.UsualPacket
	hello.Hash = proto.EMULE
	hello.Point = proto.Endpoint{Ip: 0, Port: 20033}
	hello.Properties = append(hello.Properties, proto.CreateTag(version, proto.CT_VERSION, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(capability, proto.CT_SERVER_FLAGS, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag("ged2k", proto.CT_NAME, ""))
	hello.Properties = append(hello.Properties, proto.CreateTag(versionClient, proto.CT_EMULE_VERSION, ""))

	_, serverConnection.lastError = serverConnection.SendPacket(&hello)
	log.Println("Send hello", time.Now())

	if serverConnection.lastError != nil {
		s.unregisterServerConnection <- serverConnection
		return
	}

	pc := proto.PacketCombiner{}

	for {
		ph, bytes, err := pc.Read(connection)
		if err != nil {
			serverConnection.lastError = err
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
				log.Println("Receive message from server", string(bc))
			}
		case proto.OP_SERVERSTATUS:
			ss := proto.Status{}
			ss.Get(&sb)
			if sb.Error() == nil {
				s.serverPackets <- &ss
				log.Println("Server status files:", ss.FilesCount, "users", ss.UsersCount)
			}
		case proto.OP_IDCHANGE:
			idc := proto.IdChange{}
			idc.Get(&sb)
			if sb.Error() == nil {
				log.Println("Server id change", idc.ClientId)
				s.serverPackets <- &idc
			}
		case proto.OP_SERVERIDENT:
			p := proto.UsualPacket{}
			p.Get(&sb)
			if sb.Error() == nil {
				log.Println("Received server info packet")
			}
		case proto.OP_SEARCHRESULT:
			p := proto.SearchResult{}
			p.Get(&sb)
			if sb.Error() == nil {
				log.Printf("Search result received: %d, more results %v\n", len(p.Items), p.MoreResults)
				s.serverPackets <- &p
			} else {
				log.Printf("Unable to de-serealize %v\n", sb.Error())
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
			log.Printf("Packet %x", bytes)
		}

		if sb.Error() != nil {
			serverConnection.lastError = sb.Error()
			break
		}
	}

	s.unregisterServerConnection <- serverConnection
}

func (sc *ServerConnection) SendPacket(data proto.Serializable) (int, error) {
	sz := proto.DataSize(data)
	bytes := make([]byte, sz+proto.HEADER_SIZE)
	stateBuffer := proto.StateBuffer{Data: bytes[proto.HEADER_SIZE:]}
	data.Put(&stateBuffer)

	if stateBuffer.Error() != nil {
		log.Printf("Send error %v for %d bytes\n", stateBuffer.Error(), sz)
		return 0, stateBuffer.Error()
	}

	var ph proto.PacketHeader
	bytesCount := uint32(stateBuffer.Offset() + 1)
	switch data.(type) {
	case *proto.UsualPacket:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_LOGINREQUEST}
		log.Println("Login request", sz, "bytes")
	case *proto.SearchRequest:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_SEARCHREQUEST}
		log.Printf("Search request %d bytes\n", sz)
	case *proto.SearchMore:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_QUERY_MORE_RESULT}
		log.Printf("Search more result %d bytes\n", sz)
	case *proto.GetFileSources:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_GETSOURCES}
		log.Printf("Get sources request %d bytes\n", sz)
	case *proto.GetServerList:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_GETSERVERLIST}
		log.Printf("Server list request %d bytes\n", sz)
	default:
		panic("ServerConnection Send with unknown type " + reflect.TypeOf(data).String())
	}

	sc.LastSendTime = time.Now()
	ph.Write(bytes)
	return sc.connection.Write(bytes[:stateBuffer.Offset()+proto.HEADER_SIZE])
}

func (sc *ServerConnection) Close() {
	if sc.connection != nil {
		sc.lastError = sc.connection.Close()
	}
}
