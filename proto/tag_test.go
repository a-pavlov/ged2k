package proto

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func Test_tag(t *testing.T) {
	tag1 := Tag{}
	buf := []byte{TAGTYPE_UINT16 | 0x80, 0x11, 0x0A, 0x00}
	sb := StateBuffer{Data: buf}
	sb.Read(&tag1)
	if sb.err != nil {
		t.Errorf("Tag read fail %x", sb.err)
	}

	if tag1.AsUint16() != 0x0A {
		t.Errorf("Tag value as uint16 vrong value %d", tag1.AsUint16())
	}

	{
		var v2 uint16 = 1024
		tag2 := CreateTag(v2, FT_FILESIZE, "")

		buf_exp := []byte{TAGTYPE_UINT16 | 0x80, FT_FILESIZE, 0x00, 0x00}
		binary.LittleEndian.PutUint16(buf_exp[2:], 1024)

		tag2.Put(&StateBuffer{Data: buf})

		if !bytes.Equal(buf, buf_exp) {
			t.Errorf("Wrong Tag U16 write result %x expected %x", buf, buf_exp)
		}
	}

	{
		buf_exp := []byte{TAGTYPE_UINT16 | 0x80, FT_FILESIZE, 0x00, 0x00, 0x00, 0x00}
		binary.LittleEndian.PutUint32(buf_exp[2:], 0xABABABAB)

		var v2 uint32 = 0xABABABAB
		tag2 := CreateTag(v2, FT_FILESIZE, "")

		tag2.Put(&StateBuffer{Data: buf})
		if bytes.Equal(buf, buf_exp) {
			t.Errorf("Wrong Tag U32 write result %x expected %x", buf, buf_exp)
		}
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

	buf_recv := make([]byte, len(buf))
	sb_recv := StateBuffer{Data: buf_recv}
	sb_recv.Write(sz).Write(c)

	if sb_recv.err != nil {
		t.Errorf("Can not serialize tag collection back %v", sb_recv.err)
	}

	if !bytes.Equal(buf, buf_recv) {
		t.Errorf("Wrong content %x expected %x", buf_recv, buf)
	}
}

type TagTest struct {
	t   Tag
	exp []byte
}

func Test_tagCreation(t *testing.T) {
	templates := []TagTest{
		/*1 byte*/ {t: CreateTag(uint8(0xED), 0x10, ""), exp: []byte{(TAGTYPE_UINT8 | 0x80), 0x10, 0xED}},
		/*2 bytes*/ {t: CreateTag(uint16(0x0D0A), 0x11, ""), exp: []byte{(TAGTYPE_UINT16 | 0x80), 0x11, 0x0A, 0x0D}},
		/*8 bytes*/ {t: CreateTag(uint64(0x0807060504030201), 0x00, "0123"), exp: []byte{TAGTYPE_UINT64, 0x04, 0x00, 0x30, 0x31, 0x32, 0x33, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}},
		/*variable string*/ {t: CreateTag("GOSTRING1234567890", 0x00, "ABCD"), exp: []byte{TAGTYPE_STRING, 0x04, 0x00, 'A', 'B', 'C', 'D', 0x12, 0x00, 'G', 'O', 'S', 'T', 'R', 'I', 'N', 'G', '1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}},
		/*defined string*/ {t: CreateTag("APPLE", 0x00, "IVAN"), exp: []byte{(TAGTYPE_STR5), 0x04, 0x00, 'I', 'V', 'A', 'N', 'A', 'P', 'P', 'L', 'E'}},
		/*blob*/ {t: CreateTag([]byte{0x0D, 0x0A, 0x0B}, 0x0A, ""), exp: []byte{(TAGTYPE_BLOB | 0x80), 0x0A, 0x03, 0x00, 0x00, 0x00, 0x0D, 0x0A, 0x0B}},
		/*float TagTest(TAGTYPE_FLOAT32 | 0x80), 0x15, 0x01, 0x02, 0x03, 0x04, */
		/*bool*/ {t: CreateTag(true, 0x15, ""), exp: []byte{(TAGTYPE_BOOL | 0x80), 0x15, 0x01}},
		/*hash*/ {t: CreateTag(Hash{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}, 0x20, ""), exp: []byte{(TAGTYPE_HASH16 | 0x80), 0x20, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}}}

	buf := make([]byte, 50)
	for i, test := range templates {
		sw := StateBuffer{Data: buf}
		sw.Write(test.t)
		if sw.err != nil {
			t.Errorf("Tag create test %d failed with %v", i, sw.err)
		} else {
			if !bytes.Equal(buf[0:len(test.exp)], test.exp) {
				t.Errorf("Tag create test %d wrong data %x expected %x", i, buf[0:len(test.exp)], test.exp)
			}
		}
	}
}
