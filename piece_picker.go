package main

import (
	"fmt"
	"github.com/a-pavlov/ged2k/proto"
	"log"
)

type PiecePicker struct {
	BlocksInLastPiece int
	downloadingPieces []*DownloadingPiece
	pieces            proto.BitField
}

func NewPiecePicker(pieceCount int, blocksInLastPiece int) *PiecePicker {
	return &PiecePicker{BlocksInLastPiece: blocksInLastPiece, downloadingPieces: []*DownloadingPiece{}, pieces: proto.CreateBitField(pieceCount)}
}

func (pp PiecePicker) BlocksInPiece(pieceIndex int) int {
	if pieceIndex+1 == pp.pieces.Bits() {
		return pp.BlocksInLastPiece
	}

	return proto.BLOCKS_PER_PIECE
}

func (pp PiecePicker) getDownloadingPiece(pieceIndex int) *DownloadingPiece {
	for _, x := range pp.downloadingPieces {
		if x.pieceIndex == pieceIndex {
			return x
		}
	}

	return nil
}

func (pp *PiecePicker) addDownloadingBlocks(requiredBlocksCount int, peer *Peer, endGame bool) []proto.PieceBlock {
	res := []proto.PieceBlock{}
	for _, dp := range pp.downloadingPieces {
		res = append(res, dp.PickBlock(requiredBlocksCount-len(res), peer, endGame)...)
		if len(res) == requiredBlocksCount {
			break
		}
	}

	return res
}

func (pp *PiecePicker) isEndGame() bool {
	//_, _, have := pp.piecesCount()
	//return len(pp.pieceStatus)-have-len(pp.downloadingPieces) == 0 || len(pp.downloadingPieces) > END_GAME_DOWN_PIECES_LIMIT
	return true
}

func (pp *PiecePicker) chooseNextPiece() bool {
	for i := 0; i < pp.pieces.Bits(); i++ {
		if !pp.pieces.GetBit(i) {
			pp.downloadingPieces = append(pp.downloadingPieces, NewDownloadingPiece(i, pp.BlocksInPiece(i)))
			pp.pieces.SetBit(i)
			return true
		}
	}

	return false
}

func (pp *PiecePicker) PickPieces(requiredBlocksCount int, peer *Peer) []proto.PieceBlock {
	res := pp.addDownloadingBlocks(requiredBlocksCount, peer, false)

	// for medium and fast peers in end game more re-request blocks from already downloading pieces
	if peer.Speed != PEER_SPEED_SLOW && (len(res) < requiredBlocksCount) && pp.isEndGame() {
		res = append(res, pp.addDownloadingBlocks(requiredBlocksCount-len(res), peer, true)...)
	}

	if len(res) < requiredBlocksCount && pp.chooseNextPiece() {
		fmt.Printf("Required block count %d\n", requiredBlocksCount-len(res))
		res = append(res, pp.PickPieces(requiredBlocksCount-len(res), peer)...)
	}

	return res
}

func (pp *PiecePicker) AbortBlock(block proto.PieceBlock, peer *Peer) bool {
	log.Printf("Abort block %s\n", block.ToString())
	dp := pp.getDownloadingPiece(block.PieceIndex)
	if dp != nil {
		dp.AbortBlock(block.BlockIndex, peer)
		return true
	}

	return false
}

func (pp *PiecePicker) FinishBlock(pieceBlock proto.PieceBlock) {
	p := pp.getDownloadingPiece(pieceBlock.PieceIndex)
	if p != nil {
		p.FinishBlock(pieceBlock.BlockIndex)
	} else {
		log.Printf("finish block %s not in downloading queue\n", pieceBlock.ToString())
	}
}

func (pp *PiecePicker) RemoveDownloadingPiece(pieceIndex int) bool {
	for i, x := range pp.downloadingPieces {
		if x.pieceIndex == pieceIndex {
			pp.downloadingPieces = remove(pp.downloadingPieces, i)
			pp.pieces.ClearBit(pieceIndex)
			return true
		}
	}

	return false
}

func (pp *PiecePicker) SetHave(pieceIndex int) {
	if !pp.pieces.GetBit(pieceIndex) {
		panic("set have to already finished piece")
	}

	for i, x := range pp.downloadingPieces {
		if x.pieceIndex == pieceIndex {
			if x.NumBlocks() != x.NumHave() {
				panic("set piece have when not all downloading blocks are finished")
			}
			pp.downloadingPieces = remove(pp.downloadingPieces, i)
			break
		}
	}
}

func (pp *PiecePicker) IsFinished() bool {
	return pp.pieces.Count() == pp.pieces.Bits() && len(pp.downloadingPieces) == 0
}

func (pp *PiecePicker) ApplyResumeData(atp *proto.AddTransferParameters) {
	pp.pieces = atp.Pieces
	for pieceIndex, x := range atp.DownloadedBlocks {
		pp.downloadingPieces = append(pp.downloadingPieces, NewDownloadingPieceParams(pieceIndex, x))
	}
}

func (pp *PiecePicker) GetPieces() proto.BitField {
	res := proto.CloneBitField(pp.pieces)
	for _, x := range pp.downloadingPieces {
		res.ClearBit(x.pieceIndex)
	}
	return res
}

func (pp *PiecePicker) GetDownloadedBlocks() map[int]proto.BitField {
	res := make(map[int]proto.BitField)
	for _, x := range pp.downloadingPieces {
		res[x.pieceIndex] = proto.CloneBitField(x.blocksFinished)
	}

	return res
}

func remove(s []*DownloadingPiece, i int) []*DownloadingPiece {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
