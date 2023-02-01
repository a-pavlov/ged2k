package data

import "time"

const BLOCK_STATE_NONE int = 0
const BLOCK_STATE_REQUESTED int = 1
const BLOCK_STATE_WRITING int = 2
const BLOCK_STATE_FINISHED int = 3

const PEER_SPEED_SLOW int = 0
const PEER_SPEED_MEDIUM int = 1
const PEER_SPEED_FAST int = 2

type Peer struct {
	lastConnected  time.Time
	nextConnection time.Time
	failCount      int
	speed          int
}

type Block struct {
	blockState       int
	downloadersCount int
	lastDownloader   Peer
}

type PieceBlock struct {
	pieceIndex int
	pieceBlock int
}

type DownloadingPiece struct {
	pieceIndex int
	blocks     []Block
}

func CreateDownloadingPiece(pieceIndex int, blocksCount int) DownloadingPiece {
	return DownloadingPiece{pieceIndex: pieceIndex, blocks: make([]Block, blocksCount)}
}

func (dp *DownloadingPiece) BlocksWithStateCount(state int) int {
	res := 0
	for _, x := range dp.blocks {
		if x.blockState == state {
			res++
		}
	}

	return res
}

func (dp *DownloadingPiece) PickBlock(requiredBlocksCount int, peer Peer, endGame bool) []PieceBlock {
	res := []PieceBlock{}
	// not end game mode and have no free blocks
	if !endGame && dp.BlocksWithStateCount(BLOCK_STATE_REQUESTED) == len(dp.blocks) {
		return res
	}

	for i := 0; i < len(dp.blocks) && len(res) < requiredBlocksCount; i++ {
		if dp.blocks[i].blockState == BLOCK_STATE_NONE {
			res = append(res, PieceBlock{pieceIndex: dp.pieceIndex, pieceBlock: i})
			dp.blocks[i].blockState = BLOCK_STATE_REQUESTED
			dp.blocks[i].lastDownloader = peer
			continue
		}

		if endGame && dp.blocks[i].blockState == BLOCK_STATE_REQUESTED {
			// re-request already requested blocks in end-game mode if new peer is faster than previous
			if dp.blocks[i].downloadersCount < 2 && dp.blocks[i].lastDownloader.speed < peer.speed && peer != dp.blocks[i].lastDownloader {
				dp.blocks[i].blockState = BLOCK_STATE_REQUESTED
				dp.blocks[i].lastDownloader = peer
				res = append(res, PieceBlock{pieceIndex: dp.pieceIndex, pieceBlock: i})
			}
		}
	}

	return res
}

func (dp *DownloadingPiece) AbortBlock(blockIndex int, peer Peer) {
	if blockIndex > len(dp.blocks) {
		panic("block index is out of range")
	}

	dp.blocks[blockIndex].blockState = BLOCK_STATE_NONE
	dp.blocks[blockIndex].downloadersCount--
	if dp.blocks[blockIndex].lastDownloader == peer {
		dp.blocks[blockIndex].lastDownloader = Peer{}
	}
}
