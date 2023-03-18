package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/a-pavlov/ged2k/proto"
	"golang.org/x/crypto/md4"
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
		incomingPieces:        make(map[int]*ReceivingPiece),
		Stat:                  MakeStatistics(),
	}
}

func (transfer *Transfer) AttachPeer(connection *PeerConnection) {
	transfer.policy.newConnection(connection)
	connection.transfer = transfer
}

func (transfer *Transfer) StartFinished(s *Session, atp *proto.AddTransferParameters) {
	execute := true
	if atp != nil && !atp.WantMoreData() {
		log.Printf("transfer %s is finished\n", transfer.Hash.ToString())
		// transfer finished
		s.transferChanFinished <- transfer
		s.transferChanResumeDataRead <- transfer
		for execute {
			select {
			case _, ok := <-transfer.cmdChan:
				if !ok {
					log.Println("Transfer exit requested")
					execute = false
				}
			}
		}

		s.transferChanClosed <- transfer
		return
	}
}

func (transfer *Transfer) Start(s *Session, atp *proto.AddTransferParameters) {
	execute := true
	var lastError error

	var hashSet *proto.HashSet
	file, openFileError := os.OpenFile(transfer.Filename, os.O_RDWR|os.O_CREATE, 0666)

	if openFileError != nil {
		lastError = openFileError
		s.transferChanError <- TransferError{transfer: transfer, err: fmt.Errorf("can not open file %s with error %v", transfer.Filename, openFileError)}
	} else {
		defer file.Close()
	}

	hashes := proto.HashSet{Hash: transfer.Hash, PieceHashes: make([]proto.ED2KHash, 0)}
	localFilename := proto.ByteContainer(transfer.Filename)

	var piecePicker PiecePicker

	if atp != nil {
		// restore state
		hashes = atp.Hashes // can be empty
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
						lastError = fmt.Errorf("seek in file %s to position %d failed %v", transfer.Filename, pb.Start(), err)
						s.transferChanError <- TransferError{transfer: transfer, err: lastError}
						break
					} else {
						n, err := file.Read(pendingBlock.data)
						if err != nil || n != len(pendingBlock.data) {
							lastError = fmt.Errorf("can not read block %s from file %s with error %v", pb.ToString(), transfer.Filename, err)
							s.transferChanError <- TransferError{transfer: transfer, err: lastError}
							break
						} else {
							rp.InsertBlock(&pendingBlock)
							log.Printf("%s: block %s data size: %d was restored\n", transfer.Filename, pb.ToString(), len(pendingBlock.data))
						}
					}
				}
			}
		}

		piecePicker = FromResumeData(atp)

		// report resume data read is finished
		s.transferChanResumeDataRead <- transfer
	} else {
		piecePicker = CreatePiecePicker(proto.NumPiecesAndBlocks(transfer.Size))
		// create initial add transfer parameters here
		s.transferResumeData <- proto.AddTransferParameters{
			Hashes:           hashes,
			Filename:         localFilename,
			Filesize:         transfer.Size,
			Pieces:           piecePicker.GetPieces(),
			DownloadedBlocks: piecePicker.GetDownloadedBlocks(),
		}
	}

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
			piecePicker.AbortBlock(apb.pendingBlock.block, apb.peer)
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
			if err != nil {
				lastError = err
				s.transferChanError <- TransferError{transfer: transfer, err: lastError}
				break
			}

			_, err = file.Write(pb.data)
			if err != nil {
				lastError = err
				s.transferChanError <- TransferError{transfer: transfer, err: lastError}
				break
			}

			err = file.Sync()
			if err != nil {
				lastError = err
				s.transferChanError <- TransferError{transfer: transfer, err: lastError}
				break
			}

			piecePicker.FinishBlock(pb.block)
			s.transferResumeData <- proto.AddTransferParameters{
				Hashes:           hashes,
				Filename:         localFilename,
				Filesize:         transfer.Size,
				Pieces:           piecePicker.GetPieces(),
				DownloadedBlocks: piecePicker.GetDownloadedBlocks(),
			}

			// piece completely downloaded
			if len(rp.blocks) == piecePicker.BlocksInPiece(pb.block.PieceIndex) {
				log.Println("Ready to hash")
				// check hash here
				if hashSet == nil {
					panic("hash set is nil!!")
				}

				if rp.Hash().Equals(hashSet.PieceHashes[pb.block.PieceIndex]) {
					// match
					// need to save resume data:
					log.Println("Hash match")
					piecePicker.SetHave(pb.block.PieceIndex)
				} else {
					log.Printf("Hash not match: %x expected %x\n", rp.Hash(), hashSet.PieceHashes[pb.block.PieceIndex])
					// restore piece as no-have
					piecePicker.RemoveDownloadingPiece(pb.block.PieceIndex)
				}

				s.transferResumeData <- proto.AddTransferParameters{
					Hashes:           hashes,
					Filename:         localFilename,
					Filesize:         transfer.Size,
					Pieces:           piecePicker.GetPieces(),
					DownloadedBlocks: piecePicker.GetDownloadedBlocks(),
				}

				wasFinished := piecePicker.IsFinished()
				piecePicker.SetHave(pb.block.PieceIndex)
				delete(transfer.incomingPieces, pb.block.PieceIndex)
				if !wasFinished && piecePicker.IsFinished() {
					// disconnect all peers
					// status finished
					// need save resume data
					// nothing to do - all pieces marked as downloaded
					log.Println("All data was received, close file")
					s.transferChanFinished <- transfer
				}
			}

		case peerConnection := <-transfer.peerConnChan:
			if lastError != nil {
				log.Println("Ready to download file - transfer error, close")
				peerConnection.Close(true)
				break
			}

			log.Println("Ready to download file")
			//if peerConnection.peer == nil
			blocks := piecePicker.PickPieces(proto.REQUEST_QUEUE_SIZE, peerConnection.peer)
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
