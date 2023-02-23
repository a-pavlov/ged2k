package proto

import (
	"bytes"
	"fmt"
	"testing"
)

func Test_PieceBlockSerialize(t *testing.T) {
	data := []byte{0x04, 0x00, 0x00, 0x00, 0x0F, 0x00, 0x00, 0x00}
	sb := StateBuffer{Data: data}
	pb := PieceBlock{}
	sb.Read(&pb)
	if sb.Error() != nil {
		t.Errorf("Unable to read piece block: %v", sb.Error())
	} else if pb.PieceIndex != 4 || pb.BlockIndex != 15 {
		t.Errorf("Piece index %v or block index %v are not correct", pb.PieceIndex, pb.BlockIndex)
	}

	data2 := make([]byte, 8)
	sb2 := StateBuffer{Data: data2}
	sb2.Write(pb)
	if sb2.Error() != nil {
		t.Errorf("Can not write piece block: %v", sb2.Error())
	} else if !bytes.Equal(data, data2) {
		t.Errorf("Source %v does not match res %v", data, data2)
	}
}

func Test_AddTransferParameters(t *testing.T) {
	atp_1 := AddTransferParameters{
		Hashes:           HashSet{Hash: EMULE, PieceHashes: []EMuleHash{EMULE, Terminal}},
		Filename:         String2ByteContainer("/tmp/test.data"),
		Filesize:         uint64(PIECE_SIZE * 2),
		DownloadedBlocks: make([]PieceBlock, 0)}

	atp_2 := AddTransferParameters{
		Hashes:           HashSet{Hash: EMULE, PieceHashes: []EMuleHash{EMULE, Terminal, ZERO}},
		Filename:         String2ByteContainer("/tmp/data1/data2/some_long_filename_here.data"),
		Filesize:         uint64(PIECE_SIZE * 2),
		DownloadedBlocks: []PieceBlock{{BlockIndex: 1, PieceIndex: 2}, {BlockIndex: 3, PieceIndex: 4}}}

	data := make([]byte, 300)
	sb := StateBuffer{Data: data}
	sb.Write(atp_1)
	if sb.Error() != nil {
		t.Errorf("Can not write atp 1: %v", sb.Error())
	} else if sb.Offset() != atp_1.Size() {
		t.Errorf("Wrong size on atp 1: %v expected %v", atp_1.Size(), sb.Offset())
	}

	sb.Write(atp_2)

	if sb.Error() != nil {
		t.Errorf("Unable to write atp 2: %v", sb.Error())
	}

	if sb.Offset() != atp_1.Size()+atp_2.Size() {
		t.Errorf("Wrong write size %v", atp_1.Size()+atp_2.Size())
	}

	var atp_1_r AddTransferParameters
	sb2 := StateBuffer{Data: data}
	sb2.Read(&atp_1_r)
	if sb2.Error() != nil {
		t.Errorf("Can not read atp 1: %v", sb2.Error())
	} else if sb2.Offset() != atp_1_r.Size() {
		t.Errorf("Size error on read atp 1: %v expected %v", atp_1_r.Size(), sb2.Offset())
	}

	fmt.Println("ATP1 size", atp_1_r.Size())

	var atp_2_r AddTransferParameters
	sb2.Read(&atp_2_r)
	if sb2.Error() != nil {
		t.Errorf("Can not read atps: %v", sb2.Error())
	}

	if atp_1.Size() != atp_1_r.Size() {
		t.Error("Wrong sizes on atp 1")
	}

	if !bytes.Equal(atp_1.Filename, atp_1_r.Filename) {
		t.Error("Filenames not match atp 1")
	}

	if atp_2.DownloadedBlocks[1] != atp_2_r.DownloadedBlocks[1] {
		t.Error("Downloaded blocks in atp 2 are not equal")
	}
}
