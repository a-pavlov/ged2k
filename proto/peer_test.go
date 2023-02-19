package proto

import (
	"bytes"
	"fmt"
	"testing"
)

func Test_requests(t *testing.T) {
	rp32 := RequestParts32{H: LIBED2K, BeginOffset: [PARTS_IN_REQUEST]uint32{uint32(0), uint32(1), uint32(2)}, EndOffset: [PARTS_IN_REQUEST]uint32{uint32(3), uint32(4), uint32(5)}}
	rp64 := RequestParts64{H: LIBED2K, BeginOffset: [PARTS_IN_REQUEST]uint64{uint64(6), uint64(7), uint64(8)}, EndOffset: [PARTS_IN_REQUEST]uint64{uint64(9), uint64(10), uint64(11)}}
	if DataSize(rp32) != 40 {
		t.Errorf("Size of request 32 is not correct %d\n", DataSize(rp32))
	}

	if DataSize(rp64) != 64 {
		t.Errorf("Size of request 64 is not correct %d\n", DataSize(rp64))
	}

	buf := make([]byte, 104)
	sb := StateBuffer{Data: buf}
	sb.Write(rp32).Write(rp64)
	if sb.Error() != nil {
		t.Errorf("Buffer error %v", sb.Error())
	}

	sb_in := StateBuffer{Data: buf}
	rp32_i := RequestParts32{}
	rp64_i := RequestParts64{}
	sb_in.Read(&rp32_i).Read(&rp64_i)
	if sb_in.Error() != nil {
		t.Errorf("Error on read requests %v\n", sb_in.Error())
	}

	if rp32 != rp32_i {
		t.Error("No equal 32\n")
	}

	if rp64 != rp64_i {
		t.Error("No equal 32\n")
	}
}

func Test_HashSet(t *testing.T) {
	hs := HashSet{H: EMULE, PieceHashes: []Hash{EMULE, LIBED2K, Terminal}}
	data := make([]byte, 16+2+16*3)
	sb := StateBuffer{Data: data}
	sb.Write(hs)
	if sb.Error() != nil {
		t.Errorf("Can not write hash set %v", sb.Error())
	}

	fmt.Printf("data %v", data)

	if hs.Size() != 16+2+16*3 {
		t.Errorf("Hash set size is not correct %v", hs.Size())
	}

	hs2 := HashSet{}
	sb2 := StateBuffer{Data: data}
	sb2.Read(&hs2)
	if sb2.Error() != nil {
		t.Errorf("can not read hash set")
	} else {
		if !bytes.Equal(hs2.H[:], hs.H[:]) {
			t.Errorf("Hashes are not equal")
		} else {
			if len(hs.PieceHashes) != len(hs2.PieceHashes) {
				t.Errorf("Hash sets size are not equal")
			}

			for i := 0; i < len(hs.PieceHashes); i++ {
				if !bytes.Equal(hs2.PieceHashes[i][:], hs.PieceHashes[i][:]) {
					t.Errorf("Hashes are not equal for pos %d src %v dst %v", i, hs.PieceHashes[i], hs2.PieceHashes[i])
				}
			}
		}
	}
}
