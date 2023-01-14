package main

import (
	"fmt"
	"io"
	"reflect"

	"github.com/a-pavlov/ged2k/proto"
)

func Send(writer io.Writer, data proto.Serializable) (int, error) {
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
	return writer.Write(bytes[:stateBuffer.Offset()+proto.HEADER_SIZE])
}
