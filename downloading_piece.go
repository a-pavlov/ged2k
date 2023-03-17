package main

import (
	"github.com/a-pavlov/ged2k/proto"
	"log"
)

const BLOCK_STATE_NONE int = 0
const BLOCK_STATE_REQUESTED int = 1
const BLOCK_STATE_FINISHED int = 3

type Block struct {
	downloadersCount int
	lastDownloader   *Peer
}

type DownloadingPiece struct {
	pieceIndex      int
	blocks          []Block
	blocksRequested proto.BitField
	blocksFinished  proto.BitField
}

func NewDownloadingPiece(pieceIndex int, blocksCount int) *DownloadingPiece {
	return &DownloadingPiece{pieceIndex: pieceIndex, blocks: make([]Block, blocksCount), blocksRequested: proto.CreateBitField(blocksCount), blocksFinished: proto.CreateBitField(blocksCount)}
}

func NewDownloadingPieceParams(pieceIndex int, blocksHave proto.BitField) *DownloadingPiece {
	return &DownloadingPiece{pieceIndex: pieceIndex, blocks: make([]Block, blocksHave.Bits()), blocksRequested: proto.CloneBitField(blocksHave), blocksFinished: proto.CloneBitField(blocksHave)}
}

func (dp *DownloadingPiece) FreeBlocksCount() int {
	return dp.blocksRequested.Bits() - dp.blocksRequested.Count()
}

func (dp *DownloadingPiece) IsBlockRequested(blockIndex int) bool {
	return dp.blocksRequested.GetBit(blockIndex)
}

func (dp *DownloadingPiece) IsBlockFinished(blockIndex int) bool {
	return dp.blocksFinished.GetBit(blockIndex)
}

func (dp *DownloadingPiece) NumHave() int {
	return dp.blocksFinished.Count()
}

func (db *DownloadingPiece) NumBlocks() int {
	return db.blocksFinished.Bits()
}

func (dp *DownloadingPiece) PickBlock(requiredBlocksCount int, peer *Peer, endGame bool) []proto.PieceBlock {
	res := []proto.PieceBlock{}
	// not end game mode and have no free blocks
	if !endGame && dp.FreeBlocksCount() == 0 {
		return res
	}

	for i := 0; i < len(dp.blocks) && len(res) < requiredBlocksCount; i++ {
		if !dp.IsBlockRequested(i) {
			res = append(res, proto.PieceBlock{PieceIndex: dp.pieceIndex, BlockIndex: i})
			dp.blocksRequested.SetBit(i)
			dp.blocks[i].lastDownloader = peer
			dp.blocks[i].downloadersCount++
			continue
		}

		if endGame && dp.IsBlockRequested(i) && !dp.IsBlockFinished(i) {
			// re-request already requested blocks in end-game mode if new peer is faster than previous
			if dp.blocks[i].downloadersCount < 2 && (dp.blocks[i].lastDownloader == nil || (dp.blocks[i].lastDownloader.Speed < peer.Speed && peer != dp.blocks[i].lastDownloader)) {
				dp.blocksRequested.SetBit(i)
				dp.blocks[i].lastDownloader = peer
				dp.blocks[i].downloadersCount++
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

	if dp.IsBlockFinished(blockIndex) {
		log.Printf("can not abort block %d due to finished status\n", blockIndex)
		return
	}

	dp.blocks[blockIndex].downloadersCount--
	if dp.blocks[blockIndex].downloadersCount == 0 {
		dp.blocksRequested.ClearBit(blockIndex)
	}

	log.Printf("abort block %d peer %v last downloader %v downloaders count %d\n", blockIndex, peer, dp.blocks[blockIndex].lastDownloader, dp.blocks[blockIndex].downloadersCount)
	// block can be aborted many times - check last downloader is still not nil
	if dp.blocks[blockIndex].lastDownloader != nil {
		if dp.blocks[blockIndex].lastDownloader.endpoint == peer.endpoint {
			dp.blocks[blockIndex].lastDownloader = nil
		}
	}
}

func (dp *DownloadingPiece) FinishBlock(blockIndex int) {
	if !dp.IsBlockRequested(blockIndex) {
		panic("finish not requested block")
	}

	dp.blocksFinished.SetBit(blockIndex)
	dp.blocks[blockIndex].downloadersCount--
	dp.blocks[blockIndex].lastDownloader = nil
}
