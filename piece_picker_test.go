package main

import (
	"fmt"
	"github.com/a-pavlov/ged2k/proto"
	"testing"
)

func TestPiecePicker_PickPiecesTrivial(t *testing.T) {
	pp := CreatePiecePicker(7, 4)
	peer := Peer{endpoint: proto.EndpointFromString("192.168.11.11:7899"), Speed: PEER_SPEED_SLOW}
	for i := 0; i < 101; i++ {
		blocks := pp.PickPieces(3, peer)
		//for _, x := range blocks {
		//	fmt.Printf("piece: %d block: %d ", x.pieceIndex, x.blockIndex)
		//}

		fmt.Println("")
		if len(blocks) != 3 {
			t.Errorf("Can not obtain required blocks count %d on iteration %d", len(blocks), i)
		}
	}

	blocks2 := pp.PickPieces(3, peer)
	if len(blocks2) != 1 {
		t.Errorf("Requested block count is not match 3 expected %d", len(blocks2))
	}

	pp.RemoveDownloadingPiece(PIECE_STATE_NONE, 2)
	for i := 0; i < 16; i++ {
		blocks := pp.PickPieces(3, peer)
		if len(blocks) != 3 {
			t.Errorf("After remove can not obtain required blocks count %d on iteration %d", len(blocks), i)
		}
	}

	blocks3 := pp.PickPieces(3, peer)
	if len(blocks3) != 2 {
		t.Errorf("Can not obtain 2 blocks %d", len(blocks3))
	}
}
