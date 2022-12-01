package proto

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"golang.org/x/crypto/md4"
)

func Test_trivial(t *testing.T) {
	buf := []byte{0x01, 0x00, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	var data uint16
	rd := StateBuffer{Data: buf, err: nil}
	rd.read(&data)
	if rd.err != nil {
		t.Errorf("Failed to get uint16: %v", rd.err)
	}

	if data != 1 {
		t.Errorf("Wrong value parsed %d", data)
	}

	out_buf := make([]byte, 4)
	w := StateBuffer{Data: out_buf}
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
	sw := StateBuffer{Data: buf}
	term.Put(ed2k.Put(&sw))
	if sw.err != nil {
		t.Errorf("Hash serialize error %v", sw.err)
	}

	var h4, h5 Hash
	sr := StateBuffer{Data: buf}
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
	sw := StateBuffer{Data: buf}
	bc.Put(&sw)
	if sw.err != nil {
		t.Errorf("Byte container write failed %v", sw.err)
	}

	if !bytes.Equal(buf, []byte{0x03, 0x00, 0x01, 0x02, 0x03}) {
		t.Errorf("Byte container write wrong data %x", buf)
	}

	bc2 := ByteContainer16{}
	sr := StateBuffer{Data: buf}
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
	sw2 := StateBuffer{Data: buf2}
	bc3.Put(&sw2)
	if sw2.err != nil {
		t.Errorf("Byte container 32 failed to write %v", sw2.err)
	}

	bc4 := ByteContainer32{}
	sr2 := StateBuffer{Data: buf2}
	bc4.Get(&sr2)

	if sr2.err != nil {
		t.Errorf("Byte container 32 read failed %v", sr2.err)
	}

	if !bytes.Equal(bc4, []byte{0x04, 0x05, 0x06}) {
		t.Errorf("Byte container read wrong data %x", bc4)
	}

	x := "12345"
	bcStr := ByteContainer16(x)
	buf3 := make([]byte, 7)
	bcStr.Put(&StateBuffer{Data: buf3})
	if !bytes.Equal(buf3, []byte{0x05, 0x00, 0x31, 0x32, 0x33, 0x34, 0x35}) {
		t.Errorf("String serialize to as bytec 16 failed %x", buf3)
	}

	var bcStr2 ByteContainer16
	bcStr2.Get(&StateBuffer{Data: buf3})
	if string(bcStr2) != "12345" {
		t.Errorf("String restore from BC16 failed %s", string(bcStr2))
	}
}

func Test_tag(t *testing.T) {
	tag1 := Tag{}
	buf := []byte{TAGTYPE_UINT16 | 0x80, 0x11, 0x0A, 0x00}
	sb := StateBuffer{Data: buf}
	tag1.Get(&sb)
	if sb.err != nil {
		t.Errorf("Tag read fail %x", sb.err)
	}

	if tag1.AsUint16() != 0x0A {
		t.Errorf("Tag value as uint16 vrong value %d", tag1.AsUint16())
	}

	{
		var v2 uint16 = 1024
		tag2, err := CreateTag(v2, FT_FILESIZE)
		if err != nil {
			t.Errorf("Create tag U16 failed %v", err)
		}
		buf_exp := []byte{TAGTYPE_UINT16 | 0x80, FT_FILESIZE, 0x00, 0x00}
		binary.LittleEndian.PutUint16(buf_exp[2:], 1024)

		tag2.Put(&StateBuffer{Data: buf})
		if bytes.Equal(buf, buf_exp) {
			t.Errorf("Wrong Tag U16 write result %x expected %x", buf, buf_exp)
		}
	}

	{
		buf_exp := []byte{TAGTYPE_UINT16 | 0x80, FT_FILESIZE, 0x00, 0x00, 0x00, 0x00}
		binary.LittleEndian.PutUint32(buf_exp[2:], 0xABABABAB)

		var v2 uint32 = 0xABABABAB
		tag2, err := CreateTag(v2, FT_FILESIZE)
		if err != nil {
			t.Errorf("Create tag U32 failed %v", err)
		}

		tag2.Put(&StateBuffer{Data: buf})
		if bytes.Equal(buf, buf_exp) {
			t.Errorf("Wrong Tag U32 write result %x expected %x", buf, buf_exp)
		}
	}

}