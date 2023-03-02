package main

import (
	"github.com/a-pavlov/ged2k/proto"
	"testing"
)

const TAIL uint64 = 13766

func Test_PendingBlock(t *testing.T) {
	pieceBlock := proto.PieceBlock{PieceIndex: 7, BlockIndex: 13}
	pendingBlock := MakePendingBlock(pieceBlock, proto.PIECE_SIZE_UINT64*7+proto.BLOCK_SIZE_UINT64*13+TAIL)
	if len(pendingBlock.region.Segments) != 1 {
		t.Error("Segments count not match 1")
	} else {
		if pendingBlock.region.Segments[0].Begin != proto.PIECE_SIZE_UINT64*7+proto.BLOCK_SIZE_UINT64*13 {
			t.Errorf("Pending block region start offset not match %v", pendingBlock.region.Segments[0].Begin)
		} else {
			if pendingBlock.region.Segments[0].End != proto.PIECE_SIZE_UINT64*7+proto.BLOCK_SIZE_UINT64*13+TAIL {
				t.Errorf("Pending block end offset not match %v", pendingBlock.region.Segments[0].End)
			} else {
				if pendingBlock.region.Segments[0].End-pendingBlock.region.Segments[0].Begin != TAIL {
					t.Error("Range len not match")
				}
			}
		}
	}

	pendingBlock2 := MakePendingBlock(pieceBlock, proto.PIECE_SIZE_UINT64*7+proto.BLOCK_SIZE_UINT64*13+proto.BLOCK_SIZE_UINT64)
	if pendingBlock2.region.Segments[0].End-pendingBlock2.region.Segments[0].Begin != proto.BLOCK_SIZE_UINT64 {
		t.Errorf("Region 2 length is not correct: %v", pendingBlock2.region.Segments[0].End-pendingBlock2.region.Segments[0].Begin)
	}

	pieceBlock2 := proto.PieceBlock{PieceIndex: 0, BlockIndex: 0}
	pb2 := MakePendingBlock(pieceBlock2, 4)
	if pb2.block.PieceIndex != 0 || pb2.block.BlockIndex != 0 || len(pb2.data) != 4 || pb2.region.Segments[0].End != 4 {
		t.Errorf("Pending block create error: %v", len(pb2.data))
	}
}
