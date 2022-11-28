package proto

import (
	"bytes"
	"io"
	"testing"

	"golang.org/x/crypto/md4"
)

func Test_trivial(t *testing.T) {
	buf := []byte{0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	var data UINT16
	rd := StateReader{reader: bytes.NewReader(buf), err: nil}
	data.Get(&rd)
	if rd.err != nil {
		t.Errorf("Failed to get uint16: %v", rd.err)
	}

	if data != 1 {
		t.Errorf("Wrong value parsed %d", data)
	}

	out_buf := make([]byte, 4)
	w := StateWriter{buffer: out_buf}
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
	sw := StateWriter{buffer: buf}
	term.Put(ed2k.Put(&sw))
	if sw.err != nil {
		t.Errorf("Hash serialize error %v", sw.err)
	}

	var h4, h5 Hash
	sr := StateReader{reader: bytes.NewReader(buf)}
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
