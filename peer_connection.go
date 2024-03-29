package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/a-pavlov/ged2k/data"
	"github.com/a-pavlov/ged2k/proto"
)

const (
	PEER_SPEED_SLOW = iota
	PEER_SPEED_MEDIUM
	PEER_SPEED_FAST
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

func MakePendingBlock(b proto.PieceBlock, size uint64) PendingBlock {
	begin := b.Start()
	end := Min(b.Start()+uint64(proto.BLOCK_SIZE), size)
	return PendingBlock{block: b, region: data.MakeRegion(data.Range{Begin: begin, End: end}), data: make([]byte, end-begin)}
}

func RemovePendingBlock(s []*PendingBlock, i int) []*PendingBlock {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (pb *PendingBlock) Receive(reader io.Reader, begin uint64, end uint64) (int, error) {
	inBlockOffset := proto.InBlockOffset(begin)
	chunkLen := int(end - begin)
	n, err := io.ReadFull(reader, pb.data[inBlockOffset:inBlockOffset+chunkLen])
	log.Printf("pending block [%d:%d] receive %d bytes offset %d\n", pb.block.PieceIndex, pb.block.BlockIndex, chunkLen, inBlockOffset)
	if err == nil {
		if n != chunkLen {
			return n, fmt.Errorf("not enough bytes")
		}
		pb.region.Sub(data.Range{Begin: begin, End: end})
	}

	return n, err
}

func (pb *PendingBlock) ReceiveToEof(reader io.Reader, begin uint64) (int, error) {
	inBlockOffset := proto.InBlockOffset(begin)
	n, err := io.ReadFull(reader, pb.data[inBlockOffset:])
	if err == nil {
		pb.region.Sub(data.Range{Begin: begin, End: begin + uint64(n)})
	}

	return n, err
}

type AbortPendingBlock struct {
	pendingBlock *PendingBlock
	peer         *Peer
}

type PeerConnectionPacket struct {
	Connection *PeerConnection
	Error      error
}

type StatPacket struct {
	Connection *PeerConnection
	Counter    int
}

type PeerConnection struct {
	connection      net.Conn
	transfer        *Transfer
	peer            *Peer
	Endpoint        proto.Endpoint
	Connected       bool
	DisconnectLater bool

	Stat            Statistics
	Speed           int
	requestedBlocks []*PendingBlock
	closedByRequest bool
}

func NewPeerConnection(e proto.Endpoint, transfer *Transfer, p *Peer) *PeerConnection {
	return &PeerConnection{Endpoint: e, transfer: transfer, peer: p, Stat: MakeStatistics(), requestedBlocks: make([]*PendingBlock, 0)}
}

func (peerConnection *PeerConnection) Start(s *Session) {
	log.Println("Peer connection start", peerConnection.peer.endpoint.ToString())
	if peerConnection.connection == nil {
		conn, err := net.Dial("tcp", peerConnection.peer.endpoint.ToString())
		if err != nil {
			log.Println("Can not connect", err)
			peerConnection.unregister(s, err)
			return
		}
		hello := proto.Hello{Answer: s.CreateHelloAnswer(), HashLength: byte(proto.HASH_LEN)}
		peerConnection.connection = conn
		s.registerPeerConnection <- peerConnection
		peerConnection.SendPacket(s, proto.OP_EDONKEYPROT, proto.OP_HELLO, &hello)
	}

	pc := proto.PacketCombiner{}
	var lastError error = nil

	for {
		ph, packetBytes, err := pc.Read(peerConnection.connection)

		if err != nil {
			log.Printf("Peer connection read error %v\n", err)
			lastError = err
			break
		}

		peerConnection.recvStat(s, len(packetBytes)+ph.Size()-1)

		sb := proto.StateBuffer{Data: packetBytes}

		switch {
		case ph.Packet == proto.OP_HELLO:
			log.Println("Peer connection: HELLO")
			hello := proto.Hello{}
			sb.Read(&hello)
			if sb.Error() != nil {
				lastError = sb.Error()
				break
			}
			// obtain peer information
			helloAnswer := s.CreateHelloAnswer()
			peerConnection.SendPacket(s, proto.OP_EDONKEYPROT, proto.OP_HELLOANSWER, &helloAnswer)
		case ph.Packet == proto.OP_HELLOANSWER:
			log.Println("Peer connection: HELLO_ANSWER")
			helloAnswer := proto.HelloAnswer{}
			sb.Read(&helloAnswer)
			if sb.Error() != nil {
				lastError = sb.Error()
				break
			}

			peerConnection.SendPacket(s, proto.OP_EDONKEYPROT, proto.OP_REQUESTFILENAME, &peerConnection.transfer.Hash)
		case ph.Packet == proto.OP_PUBLICIP_REQ && ph.Protocol == proto.OP_EMULEPROT:
			log.Println("Public IP request has been received")
			ep, err := proto.FromString("192.168.111.11:9999")
			if err == nil {
				ip := proto.IP(ep.Ip)
				peerConnection.SendPacket(s, proto.OP_EMULEPROT, proto.OP_PUBLICIP_ANSWER, &ip)
			}
		case ph.Packet == proto.OP_REQUESTFILENAME:
		case ph.Packet == proto.OP_REQFILENAMEANSWER:
			fa := proto.FileAnswer{}
			sb.Read(&fa)
			if sb.Error() != nil {
				lastError = sb.Error()
				break
			}

			log.Println("Received filename answer", fa.Name.ToString())
			peerConnection.SendPacket(s, proto.OP_EDONKEYPROT, proto.OP_SETREQFILEID, &peerConnection.transfer.Hash)
		case ph.Packet == proto.OP_CANCELTRANSFER:
			// cancel transfer received
			// sent OP_REQUESTFILENAME
		case ph.Packet == proto.OP_SETREQFILEID:
			// got file status request
		case ph.Packet == proto.OP_FILESTATUS:
			fs := proto.FileStatusAnswer{}
			sb.Read(&fs)
			if sb.Error() != nil {
				lastError = sb.Error()
				log.Println("Error on file status answer", sb.Error())
				break
			}

			log.Println("File status received, bits:", fs.BF.Bits(), "count", fs.BF.Count())

			if peerConnection.transfer.Size >= proto.PIECE_SIZE_UINT64 {
				peerConnection.SendPacket(s, proto.OP_EDONKEYPROT, proto.OP_HASHSETREQUEST, &peerConnection.transfer.Hash)
			} else {
				hs := proto.HashSet{Hash: peerConnection.transfer.Hash, PieceHashes: []proto.ED2KHash{peerConnection.transfer.Hash}}
				peerConnection.transfer.hashSetChan <- &hs
				peerConnection.SendPacket(s, proto.OP_EDONKEYPROT, proto.OP_STARTUPLOADREQ, &hs.Hash)
			}
		case ph.Packet == proto.OP_FILEREQANSNOFIL:
			// no file status received
			lastError = fmt.Errorf("no file answer received")
		case ph.Packet == proto.OP_HASHSETREQUEST:
			// hash set request received
		case ph.Packet == proto.OP_HASHSETANSWER:
			// got hash set answer
			hs := proto.HashSet{}
			sb.Read(&hs)
			if sb.Error() != nil {
				lastError = sb.Error()
				log.Println("Can not read hash set answer")
				break
			} else {
				log.Println("Received hash set answer")
			}

			peerConnection.transfer.hashSetChan <- &hs
			peerConnection.SendPacket(s, proto.OP_EDONKEYPROT, proto.OP_STARTUPLOADREQ, &hs.Hash)
		case ph.Packet == proto.OP_STARTUPLOADREQ:
			// receive start upload request
		case ph.Packet == proto.OP_ACCEPTUPLOADREQ:
			log.Println("received accept uploadow req")
			peerConnection.transfer.peerConnChan <- peerConnection
			// got accept upload
		case ph.Packet == proto.OP_QUEUERANKING:
			lastError = fmt.Errorf("queue ranked: %d", sb.ReadUint16())
		case ph.Packet == proto.OP_OUTOFPARTREQS:
			lastError = fmt.Errorf("out of parts")
		case ph.Packet == proto.OP_REQUESTPARTS:
			// got 32 request parts request
		case ph.Packet == proto.OP_REQUESTPARTS_I64:
			// got 64 request parts request
		case ph.Packet == proto.OP_SENDINGPART || ph.Packet == proto.OP_SENDINGPART_I64:
			sp := proto.SendingPart{Extended: ph.Packet == proto.OP_SENDINGPART_I64}
			sb.Read(&sp)
			if sb.Error() != nil {
				lastError = sb.Error()
				break
			}

			log.Println("Peer connection sending part", sp.Begin, sp.End)

			block := proto.FromOffset(sp.Begin)

			reqBlockIndex := -1
			for i, x := range peerConnection.requestedBlocks {
				if x.block == block {
					reqBlockIndex = i
					if x.data != nil {
						recvBytes, err := x.Receive(peerConnection.connection, sp.Begin, sp.End)
						if err != nil {
							log.Printf("Peer connection error on receive data: %v\n", err)
							lastError = err
							break
						} else {
							peerConnection.recvStat(s, recvBytes)
						}

						if x.region.IsEmpty() {
							peerConnection.transfer.dataChan <- x
							peerConnection.requestedBlocks = RemovePendingBlock(peerConnection.requestedBlocks, i)
							if len(peerConnection.requestedBlocks) == 0 {
								// all blocks completed
								peerConnection.transfer.peerConnChan <- peerConnection
							}
						}
					}
					break
				}
			}

			if reqBlockIndex == -1 {
				lastError = fmt.Errorf("incoming block %s has not corresponding index", block.ToString())
			}
			// got 32 sending part response
			// got 64 sending part response
		case ph.Packet == proto.OP_COMPRESSEDPART || ph.Packet == proto.OP_COMPRESSEDPART_I64:
			// got 32 compressed part response
			// got 64 compressed part response
			cp := proto.CompressedPart{Extended: ph.Packet == proto.OP_COMPRESSEDPART_I64}
			sb.Read(&cp)
			if sb.Error() != nil {
				lastError = sb.Error()
				break
			}

			fmt.Printf("Recv compressed part, offset: %d, compressed data length %d\n", cp.Offset, cp.CompressedDataLength)

			block := proto.FromOffset(cp.Offset)
			reqBlockIndex := -1
			for i, x := range peerConnection.requestedBlocks {
				if x.block == block {
					reqBlockIndex = i
					compressedData := make([]byte, cp.CompressedDataLength)
					recvBytes, err := io.ReadFull(peerConnection.connection, compressedData)
					log.Printf("compressed recv bytes %d\n", recvBytes)

					if err != nil {
						lastError = err
						break
					} else {
						peerConnection.recvStat(s, recvBytes)
					}

					b := bytes.NewReader(compressedData)
					z, err := zlib.NewReader(b)
					if err != nil {
						lastError = err
						break
					}

					defer z.Close()

					// unpack data
					recvBytes, err = x.ReceiveToEof(z, cp.Offset)
					if err != nil {
						lastError = err
						break
					} else {
						log.Printf("receive %d uncompressed bytes\n", recvBytes)
						// already taken in account above
					}

					if x.region.IsEmpty() {
						peerConnection.transfer.dataChan <- x
						peerConnection.requestedBlocks = RemovePendingBlock(peerConnection.requestedBlocks, i)
						if len(peerConnection.requestedBlocks) == 0 {
							peerConnection.transfer.peerConnChan <- peerConnection
						}
					}
					break
				}
			}

			if reqBlockIndex == -1 {
				lastError = fmt.Errorf("incoming block %s has not corresponding index", block.ToString())
			}
		case ph.Packet == proto.OP_CANCELTRANSFER:
			lastError = fmt.Errorf("cancel transfer")
		case ph.Packet == proto.OP_END_OF_DOWNLOAD:
			// got end of download response
			lastError = fmt.Errorf("end of download")
		default:
			log.Printf("Receive unknown protocol:%x packet: %x packetBytes: %d\n", ph.Protocol, ph.Packet, ph.Bytes)
		}

		if lastError != nil {
			log.Printf("stop peer connection by error %v\n", lastError)
			break
		}
	}

	peerConnection.unregister(s, lastError)
}

func (connection *PeerConnection) SendPacket(s *Session, protocol byte, packet byte, data proto.Serializable) {
	if data == nil {
		b := make([]byte, proto.HEADER_SIZE)
		ph := proto.PacketHeader{Protocol: protocol, Packet: packet, Bytes: 1}
		ph.Write(b)
		n, err := connection.connection.Write(b[:proto.HEADER_SIZE])
		if err != nil {
			connection.Close(false)
		} else {
			connection.sendStat(s, n)
		}
	}

	sz := proto.DataSize(data)
	log.Printf("Packet size calculated: %d\n", sz)
	b := make([]byte, sz+proto.HEADER_SIZE)
	stateBuffer := proto.StateBuffer{Data: b[proto.HEADER_SIZE:]}
	data.Put(&stateBuffer)

	if stateBuffer.Error() != nil {
		log.Printf("Send error %v for %d bytes\n", stateBuffer.Error(), sz)
		connection.Close(false)
	} else {
		log.Println("Wrote", stateBuffer.Offset(), "bytes")
	}

	bytesCount := uint32(stateBuffer.Offset() + 1)
	ph := proto.PacketHeader{Protocol: protocol, Packet: packet, Bytes: bytesCount}
	ph.Write(b)
	log.Printf("Bytes: %x\n", b)
	n, err := connection.connection.Write(b[:stateBuffer.Offset()+proto.HEADER_SIZE])
	if err != nil {
		log.Printf("peer connection can not write packet %v\n", err)
		connection.Close(false)
	} else {
		connection.sendStat(s, n)
	}
}

func (peerConnection *PeerConnection) Close(byRequest bool) {
	if peerConnection.Connected {
		fmt.Println("close connection")
		if peerConnection.connection == nil {
			panic("connection is nil")
		}

		err := peerConnection.connection.Close()
		if err != nil {
			log.Printf("unable to close peed connection %v", err)
		}
	} else {
		fmt.Println("no connection - set disconnect later")
		peerConnection.DisconnectLater = true
	}
}

func (peerConnection *PeerConnection) recvStat(s *Session, n int) {
	s.statReceiveChan <- StatPacket{Connection: peerConnection, Counter: n}
}

func (peerConnection *PeerConnection) sendStat(s *Session, n int) {
	s.statSendChan <- StatPacket{Connection: peerConnection, Counter: n}
}

func (peerConnection *PeerConnection) unregister(s *Session, err error) {
	for _, pb := range peerConnection.requestedBlocks {
		peerConnection.transfer.abortPendingBlockChan <- AbortPendingBlock{pendingBlock: pb, peer: peerConnection.peer}
	}
	s.unregisterPeerConnection <- PeerConnectionPacket{Connection: peerConnection, Error: err}
}

func (peerConnection *PeerConnection) RefreshSpeed() int {
	if peerConnection.transfer != nil {
		peerDRate := peerConnection.Stat.DownloadRate()
		transferDRate := peerConnection.transfer.Stat.DownloadRate()

		switch {
		case peerDRate > 512 && peerDRate > transferDRate/16:
			peerConnection.Speed = PEER_SPEED_FAST
		case peerDRate > 4096 && peerDRate > transferDRate/64:
			peerConnection.Speed = PEER_SPEED_MEDIUM
		case peerDRate < transferDRate/15 && peerConnection.Speed == PEER_SPEED_FAST:
			peerConnection.Speed = PEER_SPEED_MEDIUM
		default:
			peerConnection.Speed = PEER_SPEED_SLOW
		}
	}

	return peerConnection.Speed
}
