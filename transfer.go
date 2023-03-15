package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/a-pavlov/ged2k/proto"
	"golang.org/x/crypto/md4"
)

const (
	TRANSFER_STATUS_READ_RESUME_DATA = iota
	TRASNFER_STATUS_STAND_BY
	TRANSFER_STATUS
)

type TransferError struct {
	transfer *Transfer
	err      error
}

type Transfer struct {
	stopped  bool
	Hash     proto.ED2KHash
	Size     uint64
	Filename string

	ReadingResumeData      bool
	Paused                 bool
	Finished               bool
	Stopped                bool
	RequestSourcesNextTime time.Time
	LastError              error

	policy                Policy
	piecePicker           *PiecePicker
	cmdChan               chan string
	dataChan              chan *PendingBlock
	sourcesChan           chan proto.FoundFileSources
	peerConnChan          chan *PeerConnection
	hashSetChan           chan *proto.HashSet
	abortPendingBlockChan chan AbortPendingBlock
	incomingPieces        map[int]*ReceivingPiece

	Stat Statistics
}

func NewTransfer(hash proto.ED2KHash, filename string, size uint64) *Transfer {
	return &Transfer{
		Hash:                  hash,
		Size:                  size,
		Filename:              filename,
		cmdChan:               make(chan string),
		dataChan:              make(chan *PendingBlock, 10),
		sourcesChan:           make(chan proto.FoundFileSources),
		peerConnChan:          make(chan *PeerConnection),
		hashSetChan:           make(chan *proto.HashSet),
		abortPendingBlockChan: make(chan AbortPendingBlock),
		policy:                MakePolicy(MAX_PEER_LIST_SIZE),
		piecePicker:           NewPiecePicker(proto.NumPiecesAndBlocks(size)),
		incomingPieces:        make(map[int]*ReceivingPiece),
		Stat:                  MakeStatistics(),
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

func (transfer *Transfer) AttachPeer(connection *PeerConnection) {
	transfer.policy.newConnection(connection)
	connection.transfer = transfer
}

func (transfer *Transfer) Start(s *Session, atp *proto.AddTransferParameters) {

	var hashSet *proto.HashSet
	file, err := os.OpenFile(transfer.Filename, os.O_WRONLY|os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		s.transferChanError <- TransferError{transfer: transfer, err: fmt.Errorf("can not open file %s with error %v", transfer.Filename, err)}
		return
	}

	defer file.Close()

	piecesCount, _ := proto.NumPiecesAndBlocks(transfer.Size)
	hashes := proto.HashSet{Hash: transfer.Hash, PieceHashes: make([]proto.ED2KHash, 0)}
	downloadedBlocks := make(map[int]*proto.BitField)
	localFilename := proto.ByteContainer(transfer.Filename)
	pieces := proto.CreateBitField(piecesCount)

	if atp != nil {
		// restore state
		hashes = atp.Hashes // can be empty
		pieces = atp.Pieces // must have
		for pieceIndex, x := range atp.DownloadedBlocks {
			rp, ok := transfer.incomingPieces[pieceIndex]
			if !ok {
				rp = &ReceivingPiece{hash: md4.New(), blocks: make([]*PendingBlock, 0)}
				transfer.incomingPieces[pieceIndex] = rp
			}

			for b := 0; b < x.Bits(); b++ {
				if x.GetBit(b) {
					pb := proto.PieceBlock{PieceIndex: pieceIndex, BlockIndex: b}
					pbSize := Min(transfer.Size-pb.Start(), proto.BLOCK_SIZE_UINT64)
					pendingBlock := PendingBlock{block: pb, data: make([]byte, pbSize)}
					_, err := file.Seek(int64(pb.Start()), 0)
					if err != nil {
						s.transferChanError <- TransferError{transfer: transfer, err: fmt.Errorf("seek in file %s to position %d failed %v", transfer.Filename, pb.Start(), err)}
						return
					} else {
						n, err := file.Read(pendingBlock.data)
						if err != nil || n != len(pendingBlock.data) {
							s.transferChanError <- TransferError{transfer: transfer, err: fmt.Errorf("can not read block %s from file %s with error %v", pb.ToString(), transfer.Filename, err)}
							return
						} else {
							rp.InsertBlock(&pendingBlock)
							log.Printf("%s: block %s data size: %d was restored\n", transfer.Filename, pb.ToString(), len(pendingBlock.data))
						}
					}
				}
			}
		}

		// report resume data read is finished
		s.transferChanResumeDataRead <- transfer
	} else {
		// create initial add transfer parameters here
		s.transferResumeData <- proto.AddTransferParameters{
			Hashes:           hashes,
			Filename:         localFilename,
			Filesize:         transfer.Size,
			Pieces:           pieces,
			DownloadedBlocks: downloadedBlocks,
		}
	}

	execute := true
	log.Println("Transfer cycle in running")
	for execute {
		select {
		case _, ok := <-transfer.cmdChan:
			if !ok {
				log.Println("Transfer exit requested")
				execute = false
			}
		case hashSet = <-transfer.hashSetChan:
		case apb := <-transfer.abortPendingBlockChan:
			log.Printf("abort block %s\n", apb.pendingBlock.block.ToString())
			transfer.piecePicker.AbortBlock(apb.pendingBlock.block, apb.peer)
		case pb := <-transfer.dataChan:
			rp, ok := transfer.incomingPieces[pb.block.PieceIndex]
			if !ok {
				log.Printf("piece %d was already removed on received block %d\n", pb.block.PieceIndex, pb.block.BlockIndex)
				// incoming block was already received and the piece has been removed from incoming order
				break
			}

			if !rp.InsertBlock(pb) {
				log.Printf("piece %d already has block %d\n", pb.block.PieceIndex, pb.block.BlockIndex)
				// incoming block has been already inserted to the piece
				break
			}

			// write block to the file in advance
			_, err := file.Seek(int64(pb.block.Start()), 0)
			if err == nil {
				file.Write(pb.data)
				file.Sync()
				transfer.piecePicker.FinishBlock(pb.block)
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
					Filesize:         transfer.Size,
					Pieces:           pieces,
					DownloadedBlocks: downloadedBlocks,
				}
			} else {
				s.transferChanError <- TransferError{transfer: transfer, err: fmt.Errorf("file %s seek error %v", transfer.Filename, err)}
				execute = false
				break
			}

			// piece completely downloaded
			if len(rp.blocks) == transfer.piecePicker.BlocksInPiece(pb.block.PieceIndex) {
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
					// restore piece as no-have
					transfer.piecePicker.RemoveDownloadingPiece(pb.block.PieceIndex)
				}

				delete(downloadedBlocks, pb.block.PieceIndex)
				s.transferResumeData <- proto.AddTransferParameters{
					Hashes:           hashes,
					Filename:         localFilename,
					Filesize:         transfer.Size,
					Pieces:           pieces,
					DownloadedBlocks: downloadedBlocks,
				}

				wasFinished := transfer.piecePicker.IsFinished()
				transfer.piecePicker.SetHave(pb.block.PieceIndex)
				delete(transfer.incomingPieces, pb.block.PieceIndex)
				if !wasFinished && transfer.piecePicker.IsFinished() {
					// disconnect all peers
					// status finished
					// need save resume data
					// nothing to do - all pieces marked as downloaded
					log.Println("All data was received, close file")
					s.transferChanFinished <- transfer
				}
			}

		case peerConnection := <-transfer.peerConnChan:
			log.Println("Ready to download file")
			//if peerConnection.peer == nil
			blocks := transfer.piecePicker.PickPieces(proto.REQUEST_QUEUE_SIZE, peerConnection.peer)
			req := proto.RequestParts64{Hash: peerConnection.transfer.Hash}
			for i, x := range blocks {
				// add piece as incoming to the transfer
				if transfer.incomingPieces[x.PieceIndex] == nil {
					transfer.incomingPieces[x.PieceIndex] = &ReceivingPiece{hash: md4.New(), blocks: make([]*PendingBlock, 0)}
				}
				pb := MakePendingBlock(x, peerConnection.transfer.Size)
				peerConnection.requestedBlocks = append(peerConnection.requestedBlocks, &pb)
				req.BeginOffset[i] = pb.region.Begin()
				req.EndOffset[i] = pb.region.Segments[0].End
				log.Println("Add to request", req.BeginOffset[i], req.EndOffset[i])
			}

			if len(blocks) > 0 {
				go peerConnection.SendPacket(s, proto.OP_EMULEPROT, proto.OP_REQUESTPARTS_I64, &req)
			} else {
				log.Println("No more blocks for peer connection")
				peerConnection.Close(true)
			}
		}
	}
	s.transferChanClosed <- transfer
}

func (transfer *Transfer) Stop() {
	close(transfer.cmdChan)
}

func (transfer *Transfer) WantMorePeers() bool {
	return transfer.LastError == nil &&
		!transfer.Paused &&
		!transfer.Finished &&
		!transfer.ReadingResumeData &&
		!transfer.Stopped &&
		transfer.policy.NumConnectCandidates() > 0
}

func (transfer *Transfer) WantMoreSources(currentTime time.Time) bool {
	return transfer.LastError == nil &&
		!transfer.Paused &&
		!transfer.Finished &&
		!transfer.ReadingResumeData &&
		!transfer.Stopped &&
		(transfer.RequestSourcesNextTime.IsZero() || currentTime.After(transfer.RequestSourcesNextTime))
}
