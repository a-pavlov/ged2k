package main

import (
	"github.com/a-pavlov/ged2k/proto"
	"log"
)

const BLOCK_STATE_NONE int = 0
const BLOCK_STATE_REQUESTED int = 1
const BLOCK_STATE_FINISHED int = 3

type Block struct {
	blockState       int
	downloadersCount int
	lastDownloader   *Peer
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

func (dp *DownloadingPiece) PickBlock(requiredBlocksCount int, peer *Peer, endGame bool) []proto.PieceBlock {
	res := []proto.PieceBlock{}
	// not end game mode and have no free blocks
	if !endGame && dp.BlocksWithStateCount(BLOCK_STATE_REQUESTED) == len(dp.blocks) {
		return res
	}

	for i := 0; i < len(dp.blocks) && len(res) < requiredBlocksCount; i++ {
		if dp.blocks[i].blockState == BLOCK_STATE_NONE {
			res = append(res, proto.PieceBlock{PieceIndex: dp.pieceIndex, BlockIndex: i})
			dp.blocks[i].blockState = BLOCK_STATE_REQUESTED
			dp.blocks[i].lastDownloader = peer
			continue
		}

		if endGame && dp.blocks[i].blockState == BLOCK_STATE_REQUESTED {
			// re-request already requested blocks in end-game mode if new peer is faster than previous
			if dp.blocks[i].downloadersCount < 2 && dp.blocks[i].lastDownloader.Speed < peer.Speed && peer != dp.blocks[i].lastDownloader {
				dp.blocks[i].blockState = BLOCK_STATE_REQUESTED
				dp.blocks[i].lastDownloader = peer
				res = append(res, proto.PieceBlock{PieceIndex: dp.pieceIndex, BlockIndex: i})
			}
		}
	}

	return res
}

func (dp *DownloadingPiece) AbortBlock(blockIndex int, peer *Peer) {
	if blockIndex > len(dp.blocks) {
		panic("block index is out of range")
	}

	if dp.blocks[blockIndex].blockState == BLOCK_STATE_FINISHED {
		log.Printf("can not abort block %d due to finished status\n", blockIndex)
	}

	dp.blocks[blockIndex].blockState = BLOCK_STATE_NONE
	dp.blocks[blockIndex].downloadersCount--
	log.Printf("abort block %d peer %v last downloader %v\n", blockIndex, peer, dp.blocks[blockIndex].lastDownloader)
	// block can be aborted many times - check last downloader is still not nil
	if dp.blocks[blockIndex].lastDownloader != nil {
		if dp.blocks[blockIndex].lastDownloader.endpoint == peer.endpoint {
			dp.blocks[blockIndex].lastDownloader = nil
		}
	}
}

func (dp *DownloadingPiece) BlockFinished(blockIndex int) {

}
