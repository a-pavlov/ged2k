package proto

import (
	"bytes"
	"io"
	"testing"

	"golang.org/x/crypto/md4"
)

func Test_trivial(t *testing.T) {
	buf := []byte{0x01, 0x00, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	var data uint16
	rd := StateBuffer{data: buf, err: nil}
	rd.read(&data)
	if rd.err != nil {
		t.Errorf("Failed to get uint16: %v", rd.err)
	}

	if data != 1 {
		t.Errorf("Wrong value parsed %d", data)
	}

	out_buf := make([]byte, 4)
	w := StateBuffer{data: out_buf}
	s := Some{Ip: 1, Port: 2}
	s.Put(&w)
	if w.Error() != nil {
		t.Errorf("Unable to serialize %v", w.Error())
	}

	if w.Bytes() != 4 {
		t.Errorf("Wrote bytes is  wrong %d", w.Bytes())
	}

	s.Put(&w)
	if w.Error() == nil {
		t.Errorf("Expected error, but no error")
	}
}

func Test_hash(t *testing.T) {
	h := md4.New()
	io.WriteString(h, "test")
	io.WriteString(h, "test2")
	//fmt.Printf("MD4 %x", h.Sum(nil))

	var term Hash = Terminal
	var ed2k Hash = LIBED2K
	//fmt.Println("Hash", h2)

	var h3 Hash
	if h3 == term {
		t.Errorf("Hashes must not be equal")
	}

	buf := make([]byte, 32)
	sw := StateBuffer{data: buf}
	term.Put(ed2k.Put(&sw))
	if sw.err != nil {
		t.Errorf("Hash serialize error %v", sw.err)
	}

	var h4, h5 Hash
	sr := StateBuffer{data: buf}
	h4.Get(h5.Get(&sr))

	if sr.err != nil {
		t.Errorf("Unable to read hashes %v", sr.err)
	}

	if h5 != ed2k {
		t.Errorf("ED2K hash %x not match %x", ed2k, h5)
	}

	if h4 != term {
		t.Errorf("Term hash %x not match %x", term, h4)
	}
}

func Test_byteContainer(t *testing.T) {
	buf := make([]byte, 5)
	bc := ByteContainer16{0x01, 0x02, 0x03}
	sw := StateBuffer{data: buf}
	bc.Put(&sw)
	if sw.err != nil {
		t.Errorf("Byte container write failed %v", sw.err)
	}

	if !bytes.Equal(buf, []byte{0x03, 0x00, 0x01, 0x02, 0x03}) {
		t.Errorf("Byte container write wrong data %x", buf)
	}

	bc2 := ByteContainer16{}
	sr := StateBuffer{data: buf}
	bc2.Get(&sr)
	if sr.err != nil {
		t.Errorf("Byte container read failed %v", sr.err)
	}

	if len(bc2) != 3 {
		t.Errorf("Byte container read len wrong %d", len(bc2))
	}

	if !bytes.Equal(bc2, []byte{0x01, 0x02, 0x03}) {
		t.Errorf("Byte container read wrong data %x", bc2)
	}

	buf2 := make([]byte, 7)
	bc3 := ByteContainer32{0x04, 0x05, 0x06}
	sw2 := StateBuffer{data: buf2}
	bc3.Put(&sw2)
	if sw2.err != nil {
		t.Errorf("Byte container 32 failed to write %v", sw2.err)
	}

	bc4 := ByteContainer32{}
	sr2 := StateBuffer{data: buf2}
	bc4.Get(&sr2)

	if sr2.err != nil {
		t.Errorf("Byte container 32 read failed %v", sr2.err)
	}

	if !bytes.Equal(bc4, []byte{0x04, 0x05, 0x06}) {
		t.Errorf("Byte container read wrong data %x", bc4)
	}

}
