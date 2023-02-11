package main

import (
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/a-pavlov/ged2k/proto"
)

type ServerConnection struct {
	buffer        []byte
	connection    net.Conn
	lastAttempt   time.Time
	lastSend      time.Time
	lastReceived  time.Time
	outgoingOrder []proto.Serializable
	address       string
}

func CreateServerConnection(a string) *ServerConnection {
	return &ServerConnection{buffer: make([]byte, 200), lastAttempt: time.Time{}, lastSend: time.Time{}, outgoingOrder: make([]proto.Serializable, 0), address: a}
}

func (sc *ServerConnection) Search(s string) error {
	parsed, err := proto.BuildEntries(0, 0, 0, 0, "", "", "", 0, 0, s)
	if err != nil {
		return err
	}

	req, err := proto.PackRequest(parsed)
	if err != nil {
		return err
	}

	sc.outgoingOrder = append(sc.outgoingOrder, &req)
	sc.Send()
	return nil
}

func (sc *ServerConnection) ServerList() {
	gsl := proto.GetServerList{}
	sc.outgoingOrder = append(sc.outgoingOrder, &gsl)
	sc.Send()
}

func (sc *ServerConnection) Send() {
	if len(sc.outgoingOrder) > 0 {
		_, err := sc.SendPacket(sc.outgoingOrder[0])

		if err != nil {
			defer sc.connection.Close()
		} else {
			sc.lastSend = time.Now()
		}
	}
}

func (sc *ServerConnection) SendPacket(data proto.Serializable) (int, error) {
	sz := proto.DataSize(data)
	bytes := make([]byte, sz+proto.HEADER_SIZE)
	stateBuffer := proto.StateBuffer{Data: bytes[proto.HEADER_SIZE:]}
	data.Put(&stateBuffer)

	if stateBuffer.Error() != nil {
		fmt.Printf("Send error %v for %d bytes\n", stateBuffer.Error(), sz)
		return 0, stateBuffer.Error()
	}

	var ph proto.PacketHeader
	bytesCount := uint32(stateBuffer.Offset() + 1)
	switch data.(type) {
	case *proto.UsualPacket:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_LOGINREQUEST}
		fmt.Println("Login request", sz, "bytes")
	case *proto.SearchRequest:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_SEARCHREQUEST}
		fmt.Printf("Search request %d bytes\n", sz)
	case *proto.SearchMore:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_QUERY_MORE_RESULT}
		fmt.Printf("Search more result %d bytes\n", sz)
	case *proto.GetFileSources:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_GETSOURCES}
		fmt.Printf("Get sources request %d bytes\n", sz)
	case *proto.GetServerList:
		ph = proto.PacketHeader{Protocol: proto.OP_EDONKEYHEADER, Bytes: bytesCount, Packet: proto.OP_GETSERVERLIST}
		fmt.Printf("Server list request %d bytes\n", sz)
	default:
		panic("ServerConnection Send with unknown type " + reflect.TypeOf(data).String())
	}

	ph.Write(bytes)
	return sc.connection.Write(bytes[:stateBuffer.Offset()+proto.HEADER_SIZE])
}
