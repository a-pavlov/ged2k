package data

import "sync"

const PIECE_STATE_NONE byte = 0
const PIECE_STATE_DOWNLOADING byte = 1
const PIECE_STATE_HAVE byte = 2
const END_GAME_DOWN_PIECES_LIMIT int = 4

type PiecePicker struct {
	mutex             sync.RWMutex
	PieceCount        int
	BlocksInLastPiece int
	downloadingPieces []*DownloadingPiece
	pieceStatus       []byte
}

func CreatePiecePicker(pieceCount int, blocksInLastPiece int) PiecePicker {
	return PiecePicker{PieceCount: pieceCount, BlocksInLastPiece: blocksInLastPiece, downloadingPieces: []*DownloadingPiece{}, pieceStatus: make([]byte, pieceCount)}
}

func (pp PiecePicker) BlocksInPiece(pieceIndex int) int {
	if pp.PieceCount == pieceIndex+1 {
		return pp.BlocksInLastPiece
	}

	return BLOCKS_PER_PIECE
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

func (pp *PiecePicker) addDownloadingBlocks(requiredBlocksCount int, peer Peer, endGame bool) []PieceBlock {
	res := []PieceBlock{}
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

func (pp *PiecePicker) сhooseNextPiece() bool {
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

func (pp *PiecePicker) PickPieces(requiredBlocksCount int, peer Peer) []PieceBlock {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	res := pp.addDownloadingBlocks(requiredBlocksCount, peer, false)

	// for medium and fast peers in end game more re-request blocks from already downloading pieces
	if peer.speed != PEER_SPEED_SLOW && (len(res) < requiredBlocksCount) && pp.isEndGame() {
		res = append(res, pp.addDownloadingBlocks(requiredBlocksCount-len(res), peer, true)...)
	}

	if len(res) < requiredBlocksCount && pp.сhooseNextPiece() {
		res = append(pp.PickPieces(requiredBlocksCount-len(res), peer))
	}

	return res
}

func (pp *PiecePicker) AbortBlock(pieceIndex int, blockIndex int, peer Peer) bool {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	dp := pp.getDownloadingPiece(pieceIndex)
	if dp != nil {
		dp.AbortBlock(blockIndex, peer)
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

func remove(s []*DownloadingPiece, i int) []*DownloadingPiece {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
