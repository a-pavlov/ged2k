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

	c.Get(&sb)

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

func Test_tagCollection(t *testing.T) {
	buf := []byte{0x09, 0x00, /* 2 bytes list size*/
		/*1 byte*/ (TAGTYPE_UINT8 | 0x80), 0x10, 0xED,
		/*2 bytes*/ (TAGTYPE_UINT16 | 0x80), 0x11, 0x0A, 0x0D,
		/*8 bytes*/ TAGTYPE_UINT64, 0x04, 0x00, 0x30, 0x31, 0x32, 0x33, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		/*variable string*/ TAGTYPE_STRING, 0x04, 0x00, 'A', 'B', 'C', 'D', 0x06, 0x00, 'S', 'T', 'R', 'I', 'N', 'G',
		/*defined string*/ (TAGTYPE_STR5), 0x04, 0x00, 'I', 'V', 'A', 'N', 'A', 'P', 'P', 'L', 'E',
		/*blob*/ (TAGTYPE_BLOB | 0x80), 0x0A, 0x03, 0x00, 0x00, 0x00, 0x0D, 0x0A, 0x0B,
		/*float*/ (TAGTYPE_FLOAT32 | 0x80), 0x15, 0x01, 0x02, 0x03, 0x04,
		/*bool*/ (TAGTYPE_BOOL | 0x80), 0x15, 0x01,
		/*hash*/ (TAGTYPE_HASH16 | 0x80), 0x20, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}

	sb := StateBuffer{Data: buf}
	var sz uint16
	sb.Read(&sz)
	if sb.Error() != nil {
		t.Errorf("Can not read tag list size %v", sb.Error())
	}

	if sz != 9 {
		t.Errorf("Tag list size incorrect %d expected 9", sz)
	}

	c := Collection{}
	for i := 0; i < int(sz); i++ {
		c = append(c, &Tag{})
	}

	sb.Read(&c)
	if sb.err != nil {
		t.Errorf("Can not read tag list %v", sb.Error())
	}

	if !c[0].(*Tag).IsByte() {
		t.Errorf("Index 0 not uint16 %v", c[0].(*Tag).Type)
	}

	if c[0].(*Tag).AsByte() != 0xED {
		t.Errorf("Index 1 value incorrect %v", c[0].(*Tag).AsByte())
	}

	if !c[1].(*Tag).IsUint16() {
		t.Errorf("Index 1 not uint16 %v", c[1].(*Tag).Type)
	}

	if c[1].(*Tag).AsUint16() != 0x0D0A {
		t.Errorf("Index 1 value incorrect %v", c[1].(*Tag).AsUint16())
	}

	if !c[2].(*Tag).IsUint64() {
		t.Errorf("Index 2 not uint64 %v", c[2].(*Tag).Type)
	}

	var x uint64 = 0x0807060504030201
	if c[2].(*Tag).AsUint64() != x {
		t.Errorf("Index 2 value incorrect %v", c[2].(*Tag).AsUint64())
	}

	if !c[3].(*Tag).IsString() {
		t.Error("Index 3 is not string")
	}

	if c[3].(*Tag).Name != "ABCD" {
		t.Errorf("Index 3 name incorrect %s", c[3].(*Tag).Name)
	}

	if c[3].(*Tag).AsString() != "STRING" {
		t.Errorf("Index 3 value incorrect %s", c[3].(*Tag).AsString())
	}

	if !c[4].(*Tag).IsString() {
		t.Error("Index 4 is not string")
	}

	if c[4].(*Tag).Name != "IVAN" {
		t.Errorf("Index 4 name incorrect %s", c[4].(*Tag).Name)
	}

	if c[4].(*Tag).AsString() != "APPLE" {
		t.Errorf("Index 4 value incorrect %s", c[4].(*Tag).AsString())
	}

	if !c[5].(*Tag).IsBlob() {
		t.Error("Index 5 is not a blob")
	}

	if !bytes.Equal(c[5].(*Tag).AsBlob(), []byte{0x0D, 0x0A, 0x0B}) {
		t.Errorf("Index 5 blob value incorrect %x", c[5].(*Tag).AsBlob())
	}

	if !c[7].(*Tag).IsBool() || !c[7].(*Tag).AsBool() {
		t.Error("Index is not bool or is not true")
	}

	if !c[8].(*Tag).IsHash() {
		t.Error("Index 8 is not a hash")
	}

	if !bytes.Equal(c[8].(*Tag).AsHash(), []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}) {
		t.Errorf("Index 8 value is not correct %x", c[8].(*Tag).AsHash())
	}
}

func Test_usualPkg(t *testing.T) {

}
