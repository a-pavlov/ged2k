package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/a-pavlov/ged2k/proto"
	"golang.org/x/crypto/md4"
)

type Transfer struct {
	pause    bool
	stopped  bool
	Hash     proto.ED2KHash
	Size     uint64
	Filename string

	connections    []*PeerConnection
	policy         Policy
	piecePicker    *PiecePicker
	waitGroup      sync.WaitGroup
	cmdChan        chan string
	dataChan       chan *PendingBlock
	sourcesChan    chan proto.FoundFileSources
	peerConnChan   chan *PeerConnection
	hashSetChan    chan *proto.HashSet
	stat           Statistics
	incomingPieces map[int]*ReceivingPiece
	//addTransferParameters proto.AddTransferParameters
	lastError error
}

func NewTransfer(hash proto.ED2KHash, filename string, size uint64) *Transfer {
	return &Transfer{
		Hash:           hash,
		Size:           size,
		Filename:       filename,
		cmdChan:        make(chan string),
		dataChan:       make(chan *PendingBlock, 10),
		sourcesChan:    make(chan proto.FoundFileSources),
		peerConnChan:   make(chan *PeerConnection),
		hashSetChan:    make(chan *proto.HashSet),
		piecePicker:    CreatePiecePicker(proto.NumPiecesAndBlocks(size)),
		incomingPieces: make(map[int]*ReceivingPiece),
	}
}

func removePeerConnection(peerConnection *PeerConnection, pc []*PeerConnection) []*PeerConnection {
	var res []*PeerConnection
	for _, x := range pc {
		if x != peerConnection {
			res = append(res, x)
		}
	}

	return res
}

func (t *Transfer) IsPaused() bool {
	return t.pause
}

func (t *Transfer) AttachPeer(connection *PeerConnection) {
	t.policy.newConnection(connection)
	t.connections = append(t.connections, connection)
	connection.transfer = t
}

func (t *Transfer) Start(s *Session, atp *proto.AddTransferParameters) {
	t.waitGroup.Add(1)

	var hashSet *proto.HashSet
	file, err := os.OpenFile(t.Filename, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		// exit by error
		defer t.waitGroup.Done()
		log.Println("Error on open file", err, t.Filename)
		t.lastError = err
		s.transferChan <- t
		return
	}

	defer file.Close()
	defer t.waitGroup.Done()

	if atp != nil {
		// restore state
	}

	// fix it with restore state
	piecesCount, _ := proto.NumPiecesAndBlocks(t.Size)
	hashes := proto.HashSet{Hash: t.Hash, PieceHashes: make([]proto.ED2KHash, 0)}
	downloadedBlocks := make(map[int]*proto.BitField)
	localFilename := proto.ByteContainer(t.Filename)
	pieces := proto.CreateBitField(piecesCount)

	execute := true
	log.Println("Transfer cycle in running")
	for execute {
		select {
		case _, ok := <-t.cmdChan:
			if !ok {
				log.Println("Transfer exit requested")
				execute = false
			}
		case hashSet = <-t.hashSetChan:
		case pb := <-t.dataChan:
			rp, ok := t.incomingPieces[pb.block.PieceIndex]
			if !ok {
				// incoming block was already received and the piece has been removed from incoming order
				break
			}

			if !rp.InsertBlock(pb) {
				// incoming block has been already inserted to the piece
				break
			}

			// write block to the file in advance
			_, e := file.Seek(int64(rp.blocks[0].block.Start()), 0)
			if e == nil {
				file.Write(pb.data)
				file.Sync()
				// need to save resume data:
				_, ok := downloadedBlocks[pb.block.PieceIndex]
				if !ok {
					bf := proto.CreateBitField(proto.BLOCKS_PER_PIECE)
					downloadedBlocks[pb.block.PieceIndex] = &bf
				}

				downloadedBlocks[pb.block.PieceIndex].SetBit(pb.block.BlockIndex)
				s.transferResumeData <- proto.AddTransferParameters{
					Hashes:           hashes,
					Filename:         localFilename,
					Filesize:         t.Size,
					Pieces:           pieces,
					DownloadedBlocks: downloadedBlocks,
				}
			} else {
				log.Printf("File seek error: %v\n", e)
				// raise the file error here and stop transfer
			}

			// piece completely downloaded
			if len(rp.blocks) == t.piecePicker.BlocksInPiece(pb.block.PieceIndex) {
				log.Println("Ready to hash")
				// check hash here
				if hashSet == nil {
					panic("hash set is nil!!")
				}

				if rp.Hash().Equals(hashSet.PieceHashes[pb.block.PieceIndex]) {
					// match
					// need to save resume data:
					log.Println("Hash match")
					pieces.SetBit(pb.block.PieceIndex)
				} else {
					log.Printf("Hash not match: %x expected %x\n", rp.Hash(), hashSet.PieceHashes[pb.block.PieceIndex])
					// hash not match
				}

				delete(downloadedBlocks, pb.block.PieceIndex)
				s.transferResumeData <- proto.AddTransferParameters{
					Hashes:           hashes,
					Filename:         localFilename,
					Filesize:         t.Size,
					Pieces:           pieces,
					DownloadedBlocks: downloadedBlocks,
				}

				wasFinished := t.piecePicker.IsFinished()
				t.piecePicker.SetHave(pb.block.PieceIndex)
				delete(t.incomingPieces, pb.block.PieceIndex)
				if !wasFinished && t.piecePicker.IsFinished() {
					// disconnect all peers
					// status finished
					// need save resume data
					// nothing to do - all pieces marked as downloaded
					log.Println("All data was received, close file")
					file.Close()
					s.transferChan <- t
				}
			}

		case peerConnection := <-t.peerConnChan:
			log.Println("Ready to download file")
			blocks := t.piecePicker.PickPieces(proto.REQUEST_QUEUE_SIZE, peerConnection.peer)
			req := proto.RequestParts64{Hash: peerConnection.transfer.Hash}
			for i, x := range blocks {
				// add piece as incoming to the transfer
				if t.incomingPieces[x.PieceIndex] == nil {
					t.incomingPieces[x.PieceIndex] = &ReceivingPiece{hash: md4.New(), blocks: make([]*PendingBlock, 0)}
				}
				pb := CreatePendingBlock(x, peerConnection.transfer.Size)
				peerConnection.requestedBlocks = append(peerConnection.requestedBlocks, &pb)
				req.BeginOffset[i] = pb.region.Begin()
				req.EndOffset[i] = pb.region.Segments[0].End
				log.Println("Add to request", req.BeginOffset[i], req.EndOffset[i])
			}

			if len(blocks) > 0 {
				go peerConnection.SendPacket(proto.OP_EMULEPROT, proto.OP_REQUESTPARTS_I64, &req)
			} else {
				log.Println("No more blocks for peer connection")
				peerConnection.Close()
			}
		}
	}

	log.Println("Transfer main cycle exit")
}

func (t *Transfer) Stop() {
	t.stopped = true
	close(t.cmdChan)
	t.waitGroup.Wait()
}

/*
func (t *Transfer) ConnectToPeer(peer Peer) {
	peer.LastConnected = time.Now()
	peer.NextConnection = time.Time{}
	connection := PeerConnection{session: t.session, peer: peer, transfer: t}
	connection.peer.peerConnection = &connection
	t.connections = append(t.connections, &connection)
	//session.connections.add(c)
	connection.Connect()
}

func (t *Transfer) PeerConnectionClose(peerConnection *PeerConnection, e error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.connections = removePeerConnection(peerConnection, t.connections)
	t.policy.PeerConnectionClosed(peerConnection, e)
}

func (t *Transfer) Tick(time time.Time, s *Session) {
	if !t.pause && !t.IsFinished() {
		for i := 0; i < 3; i++ {
			peer := t.policy.FindConnectCandidate(time)
			if !peer.IsEmpty() {
				peerConnection := s.ConnectoToPeer(peer.endpoint)
				if peerConnection != nil {
					// set peer connection to peer in policy
					//t.policy.
				}
			}
		}
	}
}*/

func (t *Transfer) WantMorePeers() bool {
	return !t.pause && !t.IsFinished() && t.policy.NumConnectCandidates() > 0
}

/*
func (t *Transfer) Tick() {
	tick := time.Tick(5000 * time.Millisecond)
	execute := true
	t.waitGroup.Add(1)
	defer t.waitGroup.Done()
	peerConnectionClosedChan := make(chan *PeerConnection)

E:
	for execute {
		select {
		case peerConn := <-peerConnectionClosedChan:
			t.policy.PeerConnectionClosed(peerConn, peerConn.lastError)
			t.session.ClosePeerConnection(peerConn.endpoint)
		case _, ok := <-t.sourcesChan:
			if ok {
				// pass sources to the policy
			}
		case comm, ok := <-t.commChan:
			if !ok {
				log.Println("Transfer exit requested")
				break E
			}

			switch comm {
			case "stop":
			// stop processing
			case "pause":
				t.pause = true
			case "resume":
				t.pause = false
			}
		case <-tick:
			log.Println("Transfer tick")
			if !t.pause && !t.IsFinsihed() {
				for i := 0; i < 3; i++ {
					peer := t.policy.FindConnectCandidate(time.Now())
					if !peer.IsEmpty() {
						peerConnection := t.session.ConnectoToPeer(peer.endpoint)
						if peerConnection != nil {
							// set peer connection to peer in policy
							//t.policy.
						}
					}
				}
			}
		}
	}
}*/

func (t *Transfer) IsFinished() bool {
	return t.piecePicker.NumHave() == t.piecePicker.PiecesCount()
}

func (t *Transfer) SecondTick(duration time.Duration, s *Session) {
	for _, x := range t.connections {
		t.stat.Add(&x.stat)
	}

	s.stat.Add(&t.stat)
	t.stat.SecondTick(duration)
}

func (t *Transfer) ConnectOnePeer(time time.Time, s *Session) {
	candidate := t.policy.FindConnectCandidate(time)
	if candidate != nil {
		pc := s.GetPeerConnectionByEndpoint(candidate.endpoint)
		if pc == nil {

		}
	}
}
