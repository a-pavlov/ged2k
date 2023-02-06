package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/a-pavlov/ged2k/data"
	"github.com/a-pavlov/ged2k/proto"
	"io"
	"net"
)

type PendingBlock struct {
	block  data.PieceBlock
	data   []byte
	region data.Region
}

func removePendingBlock(s []*PendingBlock, i int) []*PendingBlock {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (pb *PendingBlock) Receive(reader io.Reader, begin uint64, end uint64) (int, error) {
	inBlockOffset, chunkLen := data.InBlockOffset(begin, end)
	receivedBytes := 0
	for receivedBytes < chunkLen {
		n, err := reader.Read(pb.data[inBlockOffset+receivedBytes : inBlockOffset+chunkLen])
		if err != nil {
			if err == io.EOF {
				receivedBytes += n
				return receivedBytes, nil
			}

			return receivedBytes, err
		}
		receivedBytes += n
	}

	pb.region.Sub(data.Range{Begin: begin, End: end})
	return receivedBytes, nil
}

func (pb *PendingBlock) ReceiveToEnd(reader io.Reader, begin uint64) (int, error) {
	pieceBlock := data.FromOffset(begin)
	inBlockOffset, _ := data.InBlockOffset(begin, begin+1)
	// end offset is piece block start + block size - in block offset
	return pb.Receive(reader, begin, pieceBlock.Start()+uint64(data.BLOCK_SIZE)-uint64(inBlockOffset))
}

func Receive(reader io.Reader, data []byte) error {
	receivedBytes := 0
	for receivedBytes < len(data) {
		n, err := reader.Read(data[receivedBytes:])
		if err != nil {
			return err
		}

		receivedBytes += n
	}

	return nil
}

type PeerConnection struct {
	connection net.Conn
	transfer   *Transfer
	session    *Session
	peer       Peer

	recvPieceIndex    int
	recvStart         uint64
	recvLength        uint64
	recvReqCompressed bool
	recvPos           uint64
	downloadQueue     []PendingBlock
}

func (connection *PeerConnection) Connect() {
	if connection.connection != nil {
		panic("peer connection alread has connection on Connect")
	}

	if connection.transfer == nil {
		panic("transfer is null on Connect")
	}

	c, err := net.Dial("tcp", connection.peer.endpoint.AsString())
	if err != nil {
		return
	}

	connection.connection = c

	// write hello packet to peer
	ha := connection.session.CreateHelloAnswer()
	connection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_HELLO, &ha)

	// continue receive data
	connection.Start()
}

func (connection *PeerConnection) Start() {
	if connection.connection == nil {
		panic("peer connection connection is nil on Start")
	}

	pc := proto.PacketCombiner{}
	var requestedBlocks []*PendingBlock

	//buffers := [][]byte{make([]byte, 100), make([]byte, 100)}

	for {
		ph, packetBytes, err := pc.Read(connection.connection)

		if err != nil {
			fmt.Printf("Can not read packetBytes from peer %v\n", err)
			break
		}

		sb := proto.StateBuffer{Data: packetBytes}

		switch {
		case ph.Packet == proto.OP_HELLO:
			hello := proto.HelloAnswer{}
			sb.Read(&hello)
			if sb.Error() != nil {
				return
			}
			// obtain peer information
			helloAnswer := connection.session.CreateHelloAnswer()
			connection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_HELLOANSWER, &helloAnswer)
			// send hello answer
			// send file request
		case ph.Packet == proto.OP_HELLOANSWER:
			helloAnswer := proto.HelloAnswer{}
			sb.Read(&helloAnswer)
			if sb.Error() != nil {
				return
			} else {
				h := proto.Hash{}
				connection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_REQUESTFILENAME, &h)
			}
			// req filename
			connection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_REQUESTFILENAME, &connection.transfer.H)

		case ph.Packet == proto.OP_REQUESTFILENAME:
		case ph.Packet == proto.OP_REQFILENAMEANSWER:
			fa := proto.FileAnswer{}
			sb.Read(&fa)
			if sb.Error() != nil {
				return
			} else {
				h := proto.Hash{}
				connection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_FILESTATUS, &h)
			}
		case ph.Packet == proto.OP_CANCELTRANSFER:
			// cancel transfer received
			// sent OP_REQUESTFILENAME
		case ph.Packet == proto.OP_SETREQFILEID:
			// got file status request
		case ph.Packet == proto.OP_FILESTATUS:
			fs := proto.FileStatusAnswer{}
			sb.Read(&fs)
			if sb.Error() != nil {
				return
			}

			h := proto.Hash{}
			connection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_HASHSETREQUEST, &h)

			// got file status ansfer
		case ph.Packet == proto.OP_FILEREQANSNOFIL:
			// no file status received
		case ph.Packet == proto.OP_HASHSETREQUEST:
			// hash set request received
		case ph.Packet == proto.OP_HASHSETANSWER:
			// got hash set answer
			h := proto.Hash{}
			sb.Read(&h)
			count := sb.ReadUint16()
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

			connection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_STARTUPLOADREQ, &h)
		case ph.Packet == proto.OP_STARTUPLOADREQ:
			// receive start upload request
		case ph.Packet == proto.OP_ACCEPTUPLOADREQ:

			// got accept upload
		case ph.Packet == proto.OP_QUEUERANKING:
			// got queue ranking
		case ph.Packet == proto.OP_OUTOFPARTREQS:
			// got out of parts
		case ph.Packet == proto.OP_REQUESTPARTS:
			// got 32 request parts request
		case ph.Packet == proto.OP_REQUESTPARTS_I64:
			// got 64 request parts request
		case ph.Packet == proto.OP_SENDINGPART || ph.Packet == proto.OP_SENDINGPART_I64:
			sp := proto.SendingPart{Extended: ph.Packet == proto.OP_SENDINGPART_I64}
			sb.Read(&sp)
			if sb.Error() != nil {
				// raise error
			}

			block := data.FromOffset(sp.Begin)

			for i, x := range requestedBlocks {
				if x.block == block {
					if x.data != nil {
						_, err := x.Receive(connection.connection, sp.Begin, sp.End)
						if err != nil {
							// raise error and close connection
						}

						if x.region.IsEmpty() {
							removePendingBlock(requestedBlocks, i)
							// pass received data to storage
						}
					}
				}
			}

			// all blocks completed
			if len(requestedBlocks) == 0 {
				// start new request
			}

			// got 32 sending part response
			// got 64 sending part response
		case ph.Packet == proto.OP_COMPRESSEDPART || ph.Packet == proto.OP_COMPRESSEDPART_I64:
			// got 32 compressed part response
			// got 64 compressed part response
			cp := proto.CompressedPart{Extended: ph.Packet == proto.OP_COMPRESSEDPART_I64}
			sb.Read(&cp)
			if sb.Error() != nil {
				// raise error
			}

			block := data.FromOffset(cp.Offset)
			for _, x := range requestedBlocks {
				if x.block == block {
					compressedData := make([]byte, cp.CompressedDataLength)
					err := Receive(connection.connection, compressedData)
					if err != nil {
						// close connection and exit
					}

					b := bytes.NewReader(compressedData)
					z, err := zlib.NewReader(b)
					if err != nil {
						return
					}

					defer z.Close()

					_, err = x.ReceiveToEnd(z, cp.Offset)
					if err != nil {
						// close connection
					}

				}
			}

		case ph.Packet == proto.OP_END_OF_DOWNLOAD:
			// got end of download response
		default:
			fmt.Printf("Receive unknown protocol:%x packet: %x packetBytes: %d\n", ph.Protocol, ph.Packet, ph.Bytes)
		}
	}
}

func (connection *PeerConnection) SendPacket(protocol byte, packet byte, data proto.Serializable) (int, error) {
	if data == nil {
		bytes := make([]byte, proto.HEADER_SIZE)
		ph := proto.PacketHeader{Protocol: protocol, Packet: packet, Bytes: 1}
		ph.Write(bytes)
		return connection.connection.Write(bytes[:proto.HEADER_SIZE])
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
	return connection.connection.Write(bytes[:stateBuffer.Offset()+proto.HEADER_SIZE])
}

func (conneection *PeerConnection) receiveCompressedData(offset uint64, compressedLength uint64, payloadSize int) {

}

func (connection *PeerConnection) receiveData(begin uint64, end uint64, compressed bool) {
	//connection.recvPieceIndex, connection.recvStart, connection.recvLength = data.BeginEnd2StartLength(begin, end)
	//connection.recvPos = 0
	//blockIndex := int(connection.recvStart / uint64(data.BLOCK_SIZE))
	//block := connection.getDownloadingBlock(connection.recvPieceIndex, blockIndex)
	//if block != nil {
	//inBlockOffset := connection.recvStart % uint64(data.BLOCK_SIZE)
	// generate slice here
	//}
}
