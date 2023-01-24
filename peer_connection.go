package main

import (
	"fmt"
	"net"

	"github.com/a-pavlov/ged2k/proto"
)

type PeerConnection struct {
	Point      proto.Endpoint
	connection net.Conn
}

func (peerConnection *PeerConnection) Connect() {

	if peerConnection.connection == nil {
		connection, err := net.Dial("tcp", peerConnection.Point.AsString())
		if err != nil {
			fmt.Println("Connect error", err)
			return
		}

		peerConnection.connection = connection
		// write hello
	}

	pc := proto.PacketCombiner{}

	for {
		ph, bytes, error := pc.Read(peerConnection.connection)

		if error != nil {
			fmt.Printf("Can not read bytes from server %v", error)
			break
		}

		sb := proto.StateBuffer{Data: bytes}

		switch ph.Packet {
		case proto.OP_HELLO:
			sb.ReadUint8()
			// send hello answer
			// send file request
		case proto.OP_HELLOANSWER:
			// sent OP_REQUESTFILENAME
		case proto.OP_SETREQFILEID:
			// got file status request
		case proto.OP_FILESTATUS:
			// got file status ansfer
		case proto.OP_FILEREQANSNOFIL:
			// no file status received
		case proto.OP_HASHSETREQUEST:
			// hash set request received
		case proto.OP_HASHSETANSWER:
			// got hash set answer
		case proto.OP_STARTUPLOADREQ:
			// receive start upload request
		case proto.OP_ACCEPTUPLOADREQ:
			// got accept upload
		case proto.OP_QUEUERANKING:
			// got queue ranking
		case proto.OP_OUTOFPARTREQS:
			// got out of parts
		case proto.OP_REQUESTPARTS:
			// got 32 request parts request
		case proto.OP_REQUESTPARTS_I64:
			// got 64 request parts request
		case proto.OP_SENDINGPART:
			// got 32 sending part response
		case proto.OP_SENDINGPART_I64:
			// got 64 sending part response
		case proto.OP_COMPRESSEDPART:
			// got 32 compressed part response
		case proto.OP_COMPRESSEDPART_I64:
			// got 64 compressed part response
		case proto.OP_END_OF_DOWNLOAD:
			// got end of download response
		default:
			fmt.Printf("Receive unknown protocol:%x packet: %x bytes: %d\n", ph.Protocol, ph.Packet, ph.Bytes)
		}
	}
}
