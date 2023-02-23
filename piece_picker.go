package main

import (
	"fmt"
	"github.com/a-pavlov/ged2k/data"
	"sync"
)

const PIECE_STATE_NONE byte = 0
const PIECE_STATE_DOWNLOADING byte = 1
const PIECE_STATE_HAVE byte = 2
const END_GAME_DOWN_PIECES_LIMIT int = 4

type PiecePicker struct {
	mutex             sync.RWMutex
	PieceCount        int // full pieces count + 1 partial
	BlocksInLastPiece int
	downloadingPieces []*DownloadingPiece
	pieceStatus       []byte
}

func CreatePiecePicker(pieceCount int, blocksInLastPiece int) PiecePicker {
	return PiecePicker{PieceCount: pieceCount, BlocksInLastPiece: blocksInLastPiece, downloadingPieces: []*DownloadingPiece{}, pieceStatus: make([]byte, pieceCount+1)}
}

func (pp PiecePicker) BlocksInPiece(pieceIndex int) int {
	if pp.PieceCount == pieceIndex+1 {
		return pp.BlocksInLastPiece
	}

	return data.BLOCKS_PER_PIECE
}

/*
func (pp *PiecePicker) MarkAsDownloading(pieceIndex int, blockIndex int) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	p := pp.getDownloadingPiece(pieceIndex)
	if p != nil {
		b := p.blocks[blockIndex]
		b.blockState = BLOCK_STATE_REQUESTED
	}
}
*/

func (pp PiecePicker) getDownloadingPiece(pieceIndex int) *DownloadingPiece {
	for _, x := range pp.downloadingPieces {
		if x.pieceIndex == pieceIndex {
			return x
		}
	}

	return nil
}

func (pp *PiecePicker) addDownloadingBlocks(requiredBlocksCount int, peer *Peer, endGame bool) []data.PieceBlock {
	res := []data.PieceBlock{}
	for _, dp := range pp.downloadingPieces {
		res = append(res, dp.PickBlock(requiredBlocksCount-len(res), peer, endGame)...)
		if len(res) == requiredBlocksCount {
			break
		}
	}

	return res
}

func (pp *PiecePicker) isEndGame() bool {
	_, _, have := pp.piecesCount()
	return len(pp.pieceStatus)-have-len(pp.downloadingPieces) == 0 || len(pp.downloadingPieces) > END_GAME_DOWN_PIECES_LIMIT
}

func (pp *PiecePicker) chooseNextPiece() bool {
	for i, x := range pp.pieceStatus {
		if x == PIECE_STATE_NONE {
			dp := CreateDownloadingPiece(i, pp.BlocksInPiece(i))
			pp.downloadingPieces = append(pp.downloadingPieces, &dp)
			pp.pieceStatus[i] = PIECE_STATE_DOWNLOADING
			return true
		}
	}

	return false
}

func (pp *PiecePicker) piecesCount() (int, int, int) {
	none := 0
	downloading := 0
	have := 0
	for _, x := range pp.pieceStatus {
		switch x {
		case PIECE_STATE_NONE:
			none++
		case PIECE_STATE_DOWNLOADING:
			downloading++
		case PIECE_STATE_HAVE:
			have++
		}
	}

	return none, downloading, have
}

func (pp *PiecePicker) PickPieces(requiredBlocksCount int, peer *Peer) []data.PieceBlock {
	pp.mutex.Lock()
	res := pp.addDownloadingBlocks(requiredBlocksCount, peer, false)

	// for medium and fast peers in end game more re-request blocks from already downloading pieces
	if peer.Speed != PEER_SPEED_SLOW && (len(res) < requiredBlocksCount) && pp.isEndGame() {
		res = append(res, pp.addDownloadingBlocks(requiredBlocksCount-len(res), peer, true)...)
	}

	if len(res) < requiredBlocksCount && pp.chooseNextPiece() {
		fmt.Printf("Required block count %d\n", requiredBlocksCount-len(res))
		pp.mutex.Unlock()
		res = append(res, pp.PickPieces(requiredBlocksCount-len(res), peer)...)
	} else {
		pp.mutex.Unlock()
	}

	return res
}

func (pp *PiecePicker) AbortBlock(block data.PieceBlock, peer *Peer) bool {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	dp := pp.getDownloadingPiece(block.PieceIndex)
	if dp != nil {
		dp.AbortBlock(block.PieceIndex, peer)
		return true
	}

	return false
}

func (pp *PiecePicker) FinishBlock(pieceIndex int, blockIndex int) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	p := pp.getDownloadingPiece(pieceIndex)
	if p != nil {
		b := p.blocks[blockIndex]
		if b.blockState == BLOCK_STATE_FINISHED {
			panic("block state already finished")
		}
		b.blockState = BLOCK_STATE_FINISHED
		p.blocks[blockIndex] = b
	} else {
		// log downloading piece was not found
	}
}

func (pp *PiecePicker) RemoveDownloadingPiece(pieceStatus byte, pieceIndex int) bool {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	for i, x := range pp.downloadingPieces {
		if x.pieceIndex == pieceIndex {
			pp.downloadingPieces = remove(pp.downloadingPieces, i)
			pp.pieceStatus[pieceIndex] = pieceStatus
			return true
		}
	}

	return false
}

func (pp *PiecePicker) PiecesCount() int {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	return len(pp.pieceStatus)
}

func (pp *PiecePicker) NumHave() int {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	res := 0
	for _, x := range pp.pieceStatus {
		if x == PIECE_STATE_HAVE {
			res++
		}
	}

	return res
}

func (pp *PiecePicker) SetHave(pieceIndex int) {
	pp.pieceStatus[pieceIndex] = PIECE_STATE_HAVE
}

func (pp *PiecePicker) IsFinished() bool {
	for _, x := range pp.pieceStatus {
		if x != PIECE_STATE_HAVE {
			return false
		}
	}

	return true
}

func remove(s []*DownloadingPiece, i int) []*DownloadingPiece {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
