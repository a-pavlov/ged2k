package main

import (
	"fmt"
	"github.com/a-pavlov/ged2k/proto"
	"testing"
)

func TestPiecePicker_PickPiecesTrivial(t *testing.T) {
	pp := NewPiecePicker(7, 4)
	peer := Peer{endpoint: proto.EndpointFromString("192.168.11.11:7899"), Speed: PEER_SPEED_SLOW}
	for i := 0; i < 101; i++ {
		blocks := pp.PickPieces(3, &peer)
		//for _, x := range blocks {
		//	fmt.Printf("piece: %d block: %d ", x.pieceIndex, x.blockIndex)
		//}

		fmt.Println("")
		if len(blocks) != 3 {
			t.Errorf("Can not obtain required blocks count %d on iteration %d", len(blocks), i)
		}
	}

	blocks2 := pp.PickPieces(3, &peer)
	if len(blocks2) != 1 {
		t.Errorf("Requested block count is not match 3 expected %d", len(blocks2))
	}

	pp.RemoveDownloadingPiece(PIECE_STATE_NONE, 2)
	for i := 0; i < 16; i++ {
		blocks := pp.PickPieces(3, &peer)
		if len(blocks) != 3 {
			t.Errorf("After remove can not obtain required blocks count %d on iteration %d", len(blocks), i)
		}
	}

	blocks3 := pp.PickPieces(3, &peer)
	if len(blocks3) != 2 {
		t.Errorf("Can not obtain 2 blocks %d", len(blocks3))
	}
}

func Test_PiecePickerLessOneBlock(t *testing.T) {
	pp := NewPiecePicker(1, 1)
	peer := Peer{endpoint: proto.EndpointFromString("192.168.11.11:7899"), Speed: PEER_SPEED_SLOW}
	blocks := pp.PickPieces(3, &peer)
	if len(blocks) != 1 {
		t.Errorf("Blocks count requested in not correct: %v", len(blocks))
	} else if blocks[0].PieceIndex != 0 || blocks[0].BlockIndex != 0 {
		t.Errorf("Requested block has incorrect piece index %d or block index %d", blocks[0].PieceIndex, blocks[0].BlockIndex)
	}

	pp.SetHave(0)
	if !pp.IsFinished() {
		t.Errorf("Piece picker was not finished")
	}

	blocks = pp.PickPieces(3, &peer)
	if len(blocks) != 0 {
		t.Errorf("Still returning blocks after finish")
	}

	if pp.BlocksInPiece(0) != 1 {
		t.Errorf("Blocks in piece error")
	}
}
