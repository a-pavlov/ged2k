package main

import (
	"github.com/a-pavlov/ged2k/proto"
	"golang.org/x/crypto/md4"
	"testing"
)

func Test_ReceivedPiece(t *testing.T) {
	rp := ReceivingPiece{hash: md4.New(), blocks: make([]*PendingBlock, 0)}

	pb0 := PendingBlock{block: proto.PieceBlock{0, 0}, data: make([]byte, 20)}
	pb1 := PendingBlock{block: proto.PieceBlock{0, 1}, data: make([]byte, 20)}
	pb2 := PendingBlock{block: proto.PieceBlock{0, 2}, data: make([]byte, 20)}
	pb3 := PendingBlock{block: proto.PieceBlock{0, 3}, data: make([]byte, 20)}
	pb4 := PendingBlock{block: proto.PieceBlock{0, 4}, data: make([]byte, 20)}
	pb5 := PendingBlock{block: proto.PieceBlock{0, 5}, data: make([]byte, 20)}

	rp.InsertBlock(&pb3)
	if rp.hashBlockIndex != 0 {
		t.Error("Hash has started before zero block received")
	}

	rp.InsertBlock(&pb1)

	if rp.hashBlockIndex != 0 {
		t.Error("Hash has started before zero block received")
	}

	if rp.blocks[0] != &pb1 || rp.blocks[1] != &pb3 {
		t.Errorf("Wrong blocks order")
	}

	rp.InsertBlock(&pb4)

	if rp.hashBlockIndex != 0 {
		t.Error("Hash has started before zero block received")
	}

	if rp.blocks[0] != &pb1 || rp.blocks[1] != &pb3 || rp.blocks[2] != &pb4 {
		t.Errorf("Wrong blocks order")
	}

	rp.InsertBlock(&pb0)

	if rp.hashBlockIndex != 2 {
		t.Errorf("Hash has started, but not correct index: %d expected 2", rp.hashBlockIndex)
	}

	if rp.blocks[0] != &pb0 || rp.blocks[1] != &pb1 || rp.blocks[2] != &pb3 || rp.blocks[3] != &pb4 {
		t.Errorf("Wrong blocks order")
	}

	rp.InsertBlock(&pb2)

	if rp.hashBlockIndex != 5 {
		t.Errorf("Next hashing index is not correct %d expected 5", rp.hashBlockIndex)
	}

	if rp.blocks[0] != &pb0 || rp.blocks[1] != &pb1 || rp.blocks[2] != &pb2 || rp.blocks[3] != &pb3 || rp.blocks[4] != &pb4 {
		t.Errorf("Wrong blocks order")
	}

	rp.InsertBlock(&pb5)

	if rp.hashBlockIndex != 6 {
		t.Errorf("Next hashing index is not correct %d expected 6", rp.hashBlockIndex)
	}

	if rp.blocks[0] != &pb0 || rp.blocks[1] != &pb1 || rp.blocks[2] != &pb2 || rp.blocks[3] != &pb3 || rp.blocks[4] != &pb4 || rp.blocks[5] != &pb5 {
		t.Errorf("Wrong blocks order")
	}
}
