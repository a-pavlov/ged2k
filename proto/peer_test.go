package proto

import "testing"

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
