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

	pp.RemoveDownloadingPiece(2)
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
	if pp.IsFinished() {
		t.Error("Wrong finished state")
	}
	peer := Peer{endpoint: proto.EndpointFromString("192.168.11.11:7899"), Speed: PEER_SPEED_SLOW}
	blocks := pp.PickPieces(3, &peer)
	if len(blocks) != 1 {
		t.Errorf("Blocks count requested in not correct: %v", len(blocks))
	} else if blocks[0].PieceIndex != 0 || blocks[0].BlockIndex != 0 {
		t.Errorf("Requested block has incorrect piece index %d or block index %d", blocks[0].PieceIndex, blocks[0].BlockIndex)
	}

	pp.FinishBlock(blocks[0])
	if pp.IsFinished() {
		t.Error("Wrong finished state")
	}
	pp.SetHave(0)
	if !pp.IsFinished() {
		t.Error("Wrong finished state")
	}
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

func Test_PiecePickerRestore(t *testing.T) {
	src := proto.CreateBitField(7)
	src.SetBit(0)
	src.SetBit(1)
	src.SetBit(6)
	pp := NewDownloadingPieceParams(14, src)
	if pp.FreeBlocksCount() != 4 {
		t.Errorf("Free blocks count is not correct %d", pp.FreeBlocksCount())
	}

	blks := []int{0, 1, 6}
	for _, b := range blks {
		if !pp.IsBlockRequested(b) || !pp.IsBlockFinished(b) {
			t.Errorf("wrong block status: %d", b)
		}
	}

	for i := 2; i < 6; i++ {
		if pp.IsBlockRequested(i) || pp.IsBlockFinished(i) {
			t.Errorf("wrong block status: %d", i)
		}
	}

	e1, _ := proto.FromString("192.156.77.3:67889")
	peer := Peer{endpoint: e1, Speed: PEER_SPEED_SLOW}
	req_1 := pp.PickBlock(3, &peer, false)
	if len(req_1) != 3 {
		t.Errorf("Requested blocks count is not correct: %d", len(req_1))
	}

	if req_1[0].BlockIndex != 2 || req_1[1].BlockIndex != 3 || req_1[2].BlockIndex != 4 {
		t.Error("Wrong block indexes")
	}

	for _, x := range req_1 {
		if !pp.IsBlockRequested(x.BlockIndex) || pp.IsBlockFinished(x.BlockIndex) {
			t.Errorf("Wrong state in downloading piece for %d requested: %v finished %v", x.BlockIndex, pp.IsBlockRequested(x.BlockIndex), pp.IsBlockFinished(x.BlockIndex))
		}
	}

	pp.AbortBlock(2, &peer)
	// block 2 aborted
	if pp.IsBlockRequested(2) || pp.IsBlockFinished(2) {
		t.Errorf("Wrong state in downloading piece for block  2 requested: %v finished %v", pp.IsBlockRequested(2), pp.IsBlockFinished(2))
	}

	e2, _ := proto.FromString("192.156.77.4:67889")
	peer2 := Peer{endpoint: e2, Speed: PEER_SPEED_MEDIUM}

	req_2 := pp.PickBlock(3, &peer2, false)

	if len(req_2) != 2 {
		t.Errorf("Request 2 size incorrect %d", len(req_2))
	}

	if req_2[0].BlockIndex != 2 || req_2[1].BlockIndex != 5 {
		t.Error("block indexes req 2 incorrect")
	}

	peer2.Speed = PEER_SPEED_FAST

	req_3 := pp.PickBlock(3, &peer2, true)
	if len(req_3) != 2 {
		t.Errorf("req 3 size incorrect %d", len(req_3))
	}

	if req_3[0].BlockIndex != 3 || req_3[1].BlockIndex != 4 {
		t.Error("block indexes req 2 incorrect")
	}

	for i := 2; i < 6; i++ {
		if !pp.IsBlockRequested(i) || pp.IsBlockFinished(i) {
			t.Errorf("Wrong state in downloading piece for %d requested: %v finished %v", i, pp.IsBlockRequested(i), pp.IsBlockFinished(i))
		}

		if pp.blocks[i].lastDownloader != &peer2 {
			t.Errorf("block %d incorrect downloader", i)
		}

		if i == 3 || i == 4 {
			if pp.blocks[i].downloadersCount != 2 {
				t.Errorf("wrong downloaders count %d", pp.blocks[i].downloadersCount)
			}
		} else {
			if pp.blocks[i].downloadersCount != 1 {
				t.Errorf("wrong downloaders count %d", pp.blocks[i].downloadersCount)
			}
		}

		pp.FinishBlock(i)
	}

	for i := 2; i < 6; i++ {
		if !pp.IsBlockRequested(i) || !pp.IsBlockFinished(i) {
			t.Errorf("Wrong state in downloading piece for %d requested: %v finished %v", i, pp.IsBlockRequested(i), pp.IsBlockFinished(i))
		}
	}

	pp.AbortBlock(3, &peer)
	if !pp.IsBlockRequested(3) || !pp.IsBlockFinished(3) {
		t.Errorf("Aborted already finished block!")
	}

	pp.FinishBlock(4)
	if !pp.IsBlockRequested(4) || !pp.IsBlockFinished(4) {
		t.Errorf("Finish error for already finished block!")
	}

	if pp.FreeBlocksCount() != 0 {
		t.Errorf("Free blocks count incorrect")
	}
}

func Test_PiecePickerMultipleDownloaders(t *testing.T) {
	src := proto.CreateBitField(3)
	pp := NewDownloadingPieceParams(11, src)

	e1, _ := proto.FromString("192.156.77.3:889")
	e2, _ := proto.FromString("192.156.77.4:1889")
	e3, _ := proto.FromString("192.156.77.5:3889")
	peer1 := Peer{endpoint: e1, Speed: PEER_SPEED_SLOW}
	peer2 := Peer{endpoint: e2, Speed: PEER_SPEED_MEDIUM}
	peer3 := Peer{endpoint: e3, Speed: PEER_SPEED_SLOW}
	req_1 := pp.PickBlock(3, &peer1, true)
	req_2 := pp.PickBlock(3, &peer2, true)
	req_3 := pp.PickBlock(3, &peer3, true)
	if len(req_1) != 3 || len(req_2) != 3 || len(req_3) != 0 {
		t.Errorf("requesting error %d %d %d", len(req_1), len(req_2), len(req_3))
	}

	for _, x := range req_1 {
		pp.AbortBlock(x.BlockIndex, &peer2)
		if !pp.IsBlockRequested(x.BlockIndex) || pp.IsBlockFinished(x.BlockIndex) {
			t.Errorf("block %d state wrong", x.BlockIndex)
		}

		if pp.blocks[x.BlockIndex].lastDownloader != nil {
			t.Errorf("block %d last downloader is not correct", x.BlockIndex)
		}

		if pp.blocks[x.BlockIndex].downloadersCount != 1 {
			t.Errorf("block %d downloaders count wrong", x.BlockIndex)
		}
	}

	for _, x := range req_1 {
		pp.AbortBlock(x.BlockIndex, &peer1)
		if pp.IsBlockRequested(x.BlockIndex) || pp.IsBlockFinished(x.BlockIndex) {
			t.Errorf("block %d state wrong", x.BlockIndex)
		}
	}
}
