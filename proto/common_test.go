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
	rd := StateBuffer{Data: buf, err: nil}
	rd.Read(&data)
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
	bc := []byte{0x01, 0x02, 0x03}
	sw := StateBuffer{Data: buf}
	sw.Write(uint16(len(bc))).Write(bc)
	if sw.err != nil {
		t.Errorf("Byte container write failed %v", sw.err)
	}

	if !bytes.Equal(buf, []byte{0x03, 0x00, 0x01, 0x02, 0x03}) {
		t.Errorf("Byte container write wrong data %x", buf)
	}

	bc2 := make([]byte, 3)
	sr := StateBuffer{Data: buf}
	var l uint16
	sr.Read(&l).Read(bc2)

	if sr.err != nil {
		t.Errorf("Byte container read failed %v", sr.err)
	}

	if len(bc2) != 3 {
		t.Errorf("Byte container read len wrong %d", len(bc2))
	}

	if !bytes.Equal(bc2, []byte{0x01, 0x02, 0x03}) {
		t.Errorf("Byte container read wrong data %x", bc2)
	}

	x := "12345"
	bcStr := []byte(x)
	buf3 := make([]byte, 7)
	sb3 := StateBuffer{Data: buf3}
	sb3.Write(uint16(len(x))).Write(bcStr)

	//bcStr.Put(&StateBuffer{Data: buf3})
	if !bytes.Equal(buf3, []byte{0x05, 0x00, 0x31, 0x32, 0x33, 0x34, 0x35}) {
		t.Errorf("String serialize to as bytec 16 failed %x", buf3)
	}

	var bcStr2 = make([]byte, 5)
	sb4 := StateBuffer{Data: buf3}

	sb4.Read(&l).Read(bcStr2)
	//bcStr2.Get(&StateBuffer{Data: buf3})
	if string(bcStr2) != "12345" {
		t.Errorf("String restore from BC16 failed %s", string(bcStr2))
	}
}

func Test_collection(t *testing.T) {
	data := []byte{0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x03, 0x00, 0x00, 0x00, 0x05, 0x00}
	var c Collection
	var sb = StateBuffer{Data: data}
	var sz uint32
	sb.Read(&sz)

	if sz != 2 {
		t.Errorf("Collection size incorrect %d", sz)
	}

	for i := 0; i < int(sz); i++ {
		c = append(c, &Endpoint{})
	}

	sb.Read(&c)

	if sb.err != nil {
		t.Errorf("Unable to read IP collection with error %v", sb.err)
	}

	if len(c) != 2 {
		t.Errorf("IP collection size incorrect %d expected 2", len(c))
	}

	if c[0].(*Endpoint).Ip != 2 && c[0].(*Endpoint).Port != 4 {
		t.Errorf("IP 1 incorrect ip/port: %d/%d", c[0].(*Endpoint).Ip, c[0].(*Endpoint).Port)
	}

	if c[1].(*Endpoint).Ip != 3 && c[1].(*Endpoint).Port != 5 {
		t.Errorf("IP 1 incorrect ip/port: %d/%d", c[1].(*Endpoint).Ip, c[1].(*Endpoint).Port)
	}

	recv_buffer := make([]byte, len(data))
	sb2 := StateBuffer{Data: recv_buffer}
	sz2 := sz
	sb2.Write(sz2).Write(c)

	if sb2.err != nil {
		t.Errorf("Unable to write collection 32 %v", sb2.err)
	}

	if !bytes.Equal(data, recv_buffer) {
		t.Errorf("Written bytes are not correct %x", recv_buffer)
	}

}

func Test_testGetters(t *testing.T) {
	buf := []byte{0x01, 0x0, 0x0, 0x0, 0x02, 0x0, 0x03, 0x0, 0x0, 0x0, 0x0, 0xFF, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	sb := StateBuffer{Data: buf}
	a1, e1 := sb.ReadUint32()
	a2, e2 := sb.ReadUint16()
	a3, e3 := sb.ReadUint8()
	a4, e4 := sb.ReadUint32()
	a5, e5 := sb.ReadUint64()

	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
		t.Errorf("Error reading primitives %v/%v/%v/%v/%v", e1, e2, e3, e4, e5)
	}

	if a1 != 1 {
		t.Errorf("Error reading a1: %d", a1)
	}

	if a2 != 2 {
		t.Errorf("Error reading a2: %d", a2)
	}

	if a3 != 3 {
		t.Errorf("Error reading a3: %d", a3)
	}

	if a4 != 0 {
		t.Errorf("Error reading a4: %d", a4)
	}

	if a5 != 0xFF {
		t.Errorf("Error reading a5: %d", a5)
	}

	if sb.err != nil {
		t.Errorf("Error on last read %v", sb.err)
	}

	_, e := sb.ReadUint8()

	if e == nil {
		t.Errorf("Expected error here")
	}
}

func Test_usualPkg(t *testing.T) {

}
