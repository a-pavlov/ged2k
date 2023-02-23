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
	block  proto.PieceBlock
	data   []byte
	region data.Region
}

func Min(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}

	return b
}

func CreatePendingBlock(b proto.PieceBlock, size uint64) PendingBlock {
	begin := b.Start()
	end := Min(b.Start()+uint64(proto.BLOCK_SIZE), size)
	return PendingBlock{block: b, region: data.MakeRegion(data.Range{Begin: begin, End: end}), data: make([]byte, end-begin)}
}

func removePendingBlock(s []*PendingBlock, i int) []*PendingBlock {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (pb *PendingBlock) Receive(reader io.Reader, begin uint64, end uint64) (int, error) {
	inBlockOffset := data.InBlockOffset(begin)
	chunkLen := int(end - begin)
	n, err := io.ReadFull(reader, pb.data[inBlockOffset:inBlockOffset+chunkLen])
	if err == nil {
		if n != chunkLen {
			return n, fmt.Errorf("not enough bytes")
		}
		pb.region.Sub(data.Range{Begin: begin, End: end})
	}

	return n, err
}

func (pb *PendingBlock) ReceiveToEof(reader io.Reader, begin uint64) (int, error) {
	inBlockOffset := data.InBlockOffset(begin)
	n, err := io.ReadFull(reader, pb.data[inBlockOffset:])
	if err == nil {
		pb.region.Sub(data.Range{Begin: begin, End: begin + uint64(n)})
	}

	return n, err
}

type PeerConnection struct {
	connection net.Conn
	transfer   *Transfer
	peer       *Peer
	endpoint   proto.Endpoint
	address    string

	lastError       error
	stat            Statistics
	requestedBlocks []*PendingBlock
	closedByRequest bool
}

func (peerConnection *PeerConnection) Start(s *Session) {
	if peerConnection.connection == nil {
		conn, err := net.Dial("tcp", peerConnection.peer.endpoint.AsString())
		if err != nil {
			s.unregisterPeerConnection <- peerConnection
			return
		}
		ha := s.CreateHelloAnswer()
		peerConnection.connection = conn
		peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_HELLO, &ha)
	}

	pc := proto.PacketCombiner{}

	for {
		ph, packetBytes, err := pc.Read(peerConnection.connection)

		if err != nil {
			peerConnection.lastError = err
			break
		}

		sb := proto.StateBuffer{Data: packetBytes}

		switch {
		case ph.Packet == proto.OP_HELLO:
			hello := proto.HelloAnswer{}
			sb.Read(&hello)
			if sb.Error() != nil {
				peerConnection.lastError = sb.Error()
				break
			}
			// obtain peer information
			helloAnswer := s.CreateHelloAnswer()
			peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_HELLOANSWER, &helloAnswer)
			// send hello answer
			// send file request
		case ph.Packet == proto.OP_HELLOANSWER:
			helloAnswer := proto.HelloAnswer{}
			sb.Read(&helloAnswer)
			if sb.Error() != nil {
				peerConnection.lastError = sb.Error()
				break
			} else {
				peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_REQUESTFILENAME, &peerConnection.transfer.Hash)
			}
			// req filename
			peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_REQUESTFILENAME, &peerConnection.transfer.Hash)

		case ph.Packet == proto.OP_REQUESTFILENAME:
		case ph.Packet == proto.OP_REQFILENAMEANSWER:
			fa := proto.FileAnswer{}
			sb.Read(&fa)
			if sb.Error() != nil {
				peerConnection.lastError = sb.Error()
				break
			} else {
				peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_FILESTATUS, &peerConnection.transfer.Hash)
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
				peerConnection.lastError = sb.Error()
				break
			}

			peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_HASHSETREQUEST, &peerConnection.transfer.Hash)
			// got file status answer
		case ph.Packet == proto.OP_FILEREQANSNOFIL:
			// no file status received
			peerConnection.lastError = fmt.Errorf("no file answer received")
			break
		case ph.Packet == proto.OP_HASHSETREQUEST:
			// hash set request received
		case ph.Packet == proto.OP_HASHSETANSWER:
			// got hash set answer
			hs := proto.HashSet{}
			sb.Read(&hs)
			if sb.Error() != nil {
				peerConnection.lastError = sb.Error()
				break
			}

			peerConnection.transfer.hashSetChan <- &hs
			peerConnection.SendPacket(proto.OP_EDONKEYPROT, proto.OP_STARTUPLOADREQ, &hs.Hash)
		case ph.Packet == proto.OP_STARTUPLOADREQ:
			// receive start upload request
		case ph.Packet == proto.OP_ACCEPTUPLOADREQ:

			// got accept upload
		case ph.Packet == proto.OP_QUEUERANKING:
			rank := sb.ReadUint16()
			fmt.Println("queue ranked", rank)
			// got queue ranking
		case ph.Packet == proto.OP_OUTOFPARTREQS:
			// got out of parts
			fmt.Println("out of parts received")
		case ph.Packet == proto.OP_REQUESTPARTS:
			// got 32 request parts request
		case ph.Packet == proto.OP_REQUESTPARTS_I64:
			// got 64 request parts request
		case ph.Packet == proto.OP_SENDINGPART || ph.Packet == proto.OP_SENDINGPART_I64:
			sp := proto.SendingPart{Extended: ph.Packet == proto.OP_SENDINGPART_I64}
			sb.Read(&sp)
			if sb.Error() != nil {
				peerConnection.lastError = sb.Error()
				break
			}

			block := proto.FromOffset(sp.Begin)

			for i, x := range peerConnection.requestedBlocks {
				if x.block == block {
					if x.data != nil {
						_, err := x.Receive(peerConnection.connection, sp.Begin, sp.End)
						if err != nil {
							// raise error and close peerConnection
						}

						if x.region.IsEmpty() {
							peerConnection.transfer.dataChan <- x
							removePendingBlock(peerConnection.requestedBlocks, i)
						}
					}

				}
			}

			// all blocks completed
			if len(peerConnection.requestedBlocks) == 0 {
				peerConnection.transfer.peerConnChan <- peerConnection
			}

			// got 32 sending part response
			// got 64 sending part response
		case ph.Packet == proto.OP_COMPRESSEDPART || ph.Packet == proto.OP_COMPRESSEDPART_I64:
			// got 32 compressed part response
			// got 64 compressed part response
			cp := proto.CompressedPart{Extended: ph.Packet == proto.OP_COMPRESSEDPART_I64}
			sb.Read(&cp)
			if sb.Error() != nil {
				peerConnection.lastError = sb.Error()
				break
			}

			block := proto.FromOffset(cp.Offset)
			for i, x := range peerConnection.requestedBlocks {
				if x.block == block {
					compressedData := make([]byte, cp.CompressedDataLength)
					_, err := io.ReadFull(peerConnection.connection, compressedData)
					if err != nil {
						// close peerConnection and exit
					}

					b := bytes.NewReader(compressedData)
					z, err := zlib.NewReader(b)
					if err != nil {
						return
					}

					defer z.Close()

					_, err = x.ReceiveToEof(z, cp.Offset)
					if err != nil {
						peerConnection.lastError = err
					}

					if x.region.IsEmpty() {
						peerConnection.transfer.dataChan <- x
						removePendingBlock(peerConnection.requestedBlocks, i)
					}

					break
				}
			}

			if len(peerConnection.requestedBlocks) == 0 {
				peerConnection.transfer.peerConnChan <- peerConnection
			}

		case ph.Packet == proto.OP_END_OF_DOWNLOAD:
			// got end of download response
		default:
			fmt.Printf("Receive unknown protocol:%x packet: %x packetBytes: %d\n", ph.Protocol, ph.Packet, ph.Bytes)
		}
	}

	s.unregisterPeerConnection <- peerConnection
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

func (peerConnection *PeerConnection) Close() {
	peerConnection.closedByRequest = true
	err := peerConnection.connection.Close()
	if err != nil {
		fmt.Printf("unable to close peed connection %v", err)
	}
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
