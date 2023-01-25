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
			ha := proto.HelloAnswer{}
			sb.Read(&ha)
			if sb.Error() != nil {
				return
			} else {
				h := proto.Hash{}
				peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_REQUESTFILENAME, &h)
			}
			// req filename

		case proto.OP_REQUESTFILENAME:
		case proto.OP_REQFILENAMEANSWER:
			fa := proto.FileAnswer{}
			sb.Read(&fa)
			if sb.Error() != nil {
				return
			} else {
				h := proto.Hash{}
				peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_FILESTATUS, &h)
			}
		case proto.OP_CANCELTRANSFER:
			// cancel transfer received
			// sent OP_REQUESTFILENAME
		case proto.OP_SETREQFILEID:
			// got file status request
		case proto.OP_FILESTATUS:
			fs := proto.FileStatusAnswer{}
			sb.Read(&fs)
			if sb.Error() != nil {
				return
			}

			h := proto.Hash{}
			peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_HASHSETREQUEST, &h)

			// got file status ansfer
		case proto.OP_FILEREQANSNOFIL:
			// no file status received
		case proto.OP_HASHSETREQUEST:
			// hash set request received
		case proto.OP_HASHSETANSWER:
			// got hash set answer
			h := proto.Hash{}
			sb.Read(&h)
			count, _ := sb.ReadUint16()
			if sb.Error() != nil {
				return
			}

			if uint32(count) > proto.MAX_ELEMS {
				//sb.err = fmt.Errorf("elements count greater than max elements %d", sz)
				return
			}

			for i := 0; i < int(count); i++ {
				hash := proto.Hash{}
				sb.Read(&hash)

				if sb.Error() != nil {
					break
				}
			}

			peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_STARTUPLOADREQ, &h)
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

func (pc *PeerConnection) SendPacket(protocol byte, packet byte, data proto.Serializable) (int, error) {
	if data == nil {
		bytes := make([]byte, proto.HEADER_SIZE)
		ph := proto.PacketHeader{Protocol: protocol, Packet: packet, Bytes: 1}
		ph.Write(bytes)
		return pc.connection.Write(bytes[:proto.HEADER_SIZE])
	}

	sz := proto.DataSize(data)
	bytes := make([]byte, sz+proto.HEADER_SIZE)
	stateBuffer := proto.StateBuffer{Data: bytes[proto.HEADER_SIZE:]}
	data.Put(&stateBuffer)

	if stateBuffer.Error() != nil {
		fmt.Printf("Send error %v for %d bytes\n", stateBuffer.Error(), sz)
		return 0, stateBuffer.Error()
	}

	bytesCount := uint32(stateBuffer.Offset() + 1)
	ph := proto.PacketHeader{Protocol: protocol, Packet: packet, Bytes: bytesCount}
	ph.Write(bytes)
	return pc.connection.Write(bytes[:stateBuffer.Offset()+proto.HEADER_SIZE])
}
