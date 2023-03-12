package proto

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
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

	if w.Offset() != 4 {
		t.Errorf("Wrote bytes is  wrong %d", w.Offset())
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

	arr := make([]byte, 0)
	arr = h.Sum(arr)

	if len(arr) != 16 {
		t.Error("ED2KHash obtain wrong size")
	}

	//t.Errorf("ED2KHash calculated %x", arr)

	var term ED2KHash = Terminal
	var ed2k ED2KHash = LIBED2K

	var h3 ED2KHash
	if h3 == term {
		t.Errorf("Hashes must not be equal")
	}

	buf := make([]byte, 32)
	sw := StateBuffer{Data: buf}
	term.Put(ed2k.Put(&sw))
	if sw.err != nil {
		t.Errorf("ED2KHash serialize error %v", sw.err)
	}

	var h4, h5 ED2KHash
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

func Test_Hash2(t *testing.T) {
	var h1 ED2KHash
	var h2 ED2KHash
	var h3 ED2KHash
	h1 = EMULE
	h2 = EMULE
	h3 = LIBED2K
	if !h1.Equals(h2) {
		t.Errorf("Hashes are not equal: %v != %v", h1, h2)
	}

	if h1.Equals(h3) || h2.Equals(h3) {
		t.Errorf("Hashes are equal: %v = %v = %v", h1, h2, h3)
	}
}

func Test_pieceHash(t *testing.T) {
	size := 9727000
	blockSize := 190 * 1024
	offset := 0
	data := make([]byte, blockSize)
	log.Println("Started hash")
	blocksProcessed := 0
	hash := md4.New()
	for offset < size {
		currBlockSize := Min(size-offset, len(data))
		hash.Write(data[:currBlockSize])
		log.Println("ED2KHash for bytes", currBlockSize)
		offset += currBlockSize
		blocksProcessed++
	}

	if blocksProcessed != 50 {
		t.Errorf("Wrong blocks processed number %d", blocksProcessed)
	}

	arr := make([]byte, 0)
	var h ED2KHash
	hash.Sum(h[:0])
	arr = hash.Sum(arr)

	if len(arr) != 16 {
		t.Errorf("ED2KHash size is not corrrect: %d", len(arr))
	}

	expected := []byte{0x79, 0x32, 0x4E, 0x42, 0x16, 0x02, 0x25, 0x1A, 0x39, 0x93, 0x6D, 0x7E, 0x8B, 0xC6, 0x41, 0x25}
	if !bytes.Equal(expected, arr) {
		t.Errorf("Incorrect resulted hash: %v expected %v", arr, expected)
	}

	if !bytes.Equal(h[:], expected) {
		t.Errorf("Hash content not equal the expected bytes: %x != %x", h[:], expected)
	}

}

func Test_HashStrings(t *testing.T) {
	h := String2Hash("31D6CFE0D16AE931B73C59D7E0C089C0")
	if !h.Equals(Terminal) {
		t.Errorf("From string hash is not terminal %v", h)
	}

	if h.ToString() != "31D6CFE0D16AE931B73C59D7E0C089C0" {
		t.Errorf("ED2KHash format to string is not correct %v", h.ToString())
	}

	if EMULE.ToString() != "31D6CFE0D10EE931B73C59D7E0C06FC0" {
		t.Errorf("EMULE to string error %v", EMULE.ToString())
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
	a1 := sb.ReadUint32()
	a2 := sb.ReadUint16()
	a3 := sb.ReadUint8()
	a4 := sb.ReadUint32()
	a5 := sb.ReadUint64()

	if sb.Error() != nil {
		t.Errorf("Error reading primitives %v", sb.Error())
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

	sb.ReadUint8()

	if sb.Error() == nil {
		t.Errorf("Expected error here")
	}
}

func Test_usualPkg(t *testing.T) {

}

type CustomProvider struct {
	data     []byte
	portions []int
	cindx    int
}

func (cp *CustomProvider) Read(b []byte) (n int, err error) {
	startPos := 0
	endPos := len(cp.data)

	if cp.cindx > len(cp.portions) {
		return 0, io.EOF
	}

	if cp.cindx > 0 {
		startPos = cp.portions[cp.cindx-1]
	}

	if cp.cindx < len(cp.portions) {
		endPos = cp.portions[cp.cindx]
	}

	if endPos-startPos > len(b) {
		panic(fmt.Sprintf("Incoming buffer too small %d required %d", len(b), endPos-startPos))
	}

	count := copy(b, cp.data[startPos:endPos])
	cp.cindx++
	return count, nil
}

type OneByteProvider struct {
	data []byte
	pos  int
}

func (cp *OneByteProvider) Read(b []byte) (n int, err error) {
	if cp.pos >= len(cp.data) {
		return 0, io.EOF
	}

	b[0] = cp.data[cp.pos]
	cp.pos++
	return 1, nil
}

func Test_customProvider(t *testing.T) {
	cp := CustomProvider{data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
		portions: []int{2, 4, 7}}
	buf := make([]byte, 4)
	c, _ := cp.Read(buf)
	if c != 2 {
		t.Errorf("Wrong bytes provided %d expected 2", c)
	}

	if !bytes.Equal(buf[0:2], cp.data[0:2]) {
		t.Errorf("Wrong content %x stage 1", buf)
	}

	c, _ = cp.Read(buf)
	if c != 2 {
		t.Errorf("Wrong bytes provided %d expected 2", c)
	}

	if !bytes.Equal(buf[0:2], cp.data[2:4]) {
		t.Errorf("Wrong content %x stage 2", buf)
	}

	c, _ = cp.Read(buf)
	if c != 3 {
		t.Errorf("Wrong bytes provided %d expected 3", c)
	}

	if !bytes.Equal(buf[0:3], cp.data[4:7]) {
		t.Errorf("Wrong content %x stage 3", buf)
	}

	c, _ = cp.Read(buf)
	if c != 3 {
		t.Errorf("Wrong bytes provided %d expected 3", c)
	}

	if !bytes.Equal(buf[0:3], cp.data[7:10]) {
		t.Errorf("Wrong content %x stage 4", buf)
	}
}

type BufferProvider struct {
	indx int
	bufs [][]byte
}

func (bp *BufferProvider) Read(b []byte) (n int, err error) {
	if bp.indx == len(bp.bufs) {
		return 0, io.EOF
	}

	if len(bp.bufs[bp.indx]) > len(b) {
		panic(fmt.Sprintf("Incoming buffer size %d less than data size %d", len(b), len(bp.bufs[bp.indx])))
	}

	length := copy(b, bp.bufs[bp.indx])
	bp.indx++
	return length, nil
}

func Test_packetCombiner(t *testing.T) {
	cp := BufferProvider{bufs: [][]byte{{OP_EDONKEYHEADER}, {0x04, 0x00, 0x00, 0x00, OP_LOGINREQUEST}, {0x01, 0x02, 0x03}, // packet 1
		{OP_EDONKEYPROT, 0x01, 0x00, 0x00, 0x00, OP_REJECT},                                                   // packet 2
		{OP_EDONKEYHEADER, 0x07}, {0x00, 0x00, 0x00, OP_GETSERVERLIST}, {0x04, 0x05, 0x06, 0x07, 0x08, 0x09}}} // packet 3
	pc := PacketCombiner{}
	ph, data, err := pc.Read(&cp)

	expected := [][]byte{
		{0x01, 0x02, 0x03},
		{},
		{0x04, 0x05, 0x06, 0x07, 0x08, 0x09}}

	if err != nil {
		t.Errorf("Reading packet 1 error %v", err)
	} else {
		if !bytes.Equal(data, expected[0]) || ph.Packet != OP_LOGINREQUEST {
			t.Errorf("Reading packet 1 content is wrong %x expected %x", data, expected[0])
		}
	}

	ph, data, err = pc.Read(&cp)
	if err != nil {
		t.Errorf("Reading packet 2 error %v", err)
	} else if !bytes.Equal(data, expected[1]) || ph.Packet != OP_REJECT {
		t.Errorf("Reading packet 2 content is wrong %x expected %x", data, expected[1])
	}

	ph, data, err = pc.Read(&cp)
	if err != nil {
		t.Errorf("Reading packet 3 error %v", err)
	} else if !bytes.Equal(data, expected[2]) || ph.Packet != OP_GETSERVERLIST {
		t.Errorf("Reading packet 3 content is wrong %x expected %x", data, expected[2])
	}
}

/*
func Test_bp(t *testing.T) {
	cp := OneByteProvider{data: []byte{OP_EDONKEYHEADER, 0x04, 0x00, 0x00, 0x00, OP_LOGINREQUEST, 0x01, 0x02, 0x03, // packet 1
		OP_EDONKEYPROT, 0x01, 0x00, 0x00, 0x00, OP_REJECT, // packet 2
		OP_EDONKEYHEADER, 0x07, 0x00, 0x00, 0x00, OP_GETSERVERLIST, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}} // packet 3
	b, e := ReadAll(&cp)
	if e != nil {
		t.Errorf("Buffer read error %v", e)
	} else {
		t.Errorf("Length %d", len(b))
	}
} */

/*

func Test_packetCombinerRealloc(t *testing.T) {
	cp := OneByteProvider{data: []byte{OP_EDONKEYHEADER, 0x04, 0x00, 0x00, 0x00, OP_LOGINREQUEST, 0x01, 0x02, 0x03, // packet 1
		OP_EDONKEYPROT, 0x01, 0x00, 0x00, 0x00, OP_REJECT, // packet 2
		OP_EDONKEYHEADER, 0x07, 0x00, 0x00, 0x00, OP_GETSERVERLIST, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}} // packet 3

	pc := PacketCombiner{data: make([]byte, 6)}
	data, err := pc.Read(&cp)

	if err != nil {
		t.Errorf("Reading packet 1 error %v", err)
	} else {
		if !bytes.Equal(data, cp.data[0:9]) {
			t.Errorf("Reading packet 1 content is wrong %x expected %x", data, cp.data[0:9])
		}
	}

	data, err = pc.Read(&cp)
	if err != nil {
		t.Errorf("Reading packet 2 error %v", err)
	} else if !bytes.Equal(data, cp.data[9:15]) {
		t.Errorf("Reading packet 2 content is wrong %x expected %x", data, cp.data[9:15])
	}

	data, err = pc.Read(&cp)
	if err != nil {
		t.Errorf("Reading packet 3 error %v", err)
	} else if !bytes.Equal(data, cp.data[15:]) {
		t.Errorf("Reading packet 3 content is wrong %x expected %x", data, cp.data[15:])
	}

	if len(pc.data) != 12 { // incomig buffer x2 is enougn
		t.Errorf("Buffer has wrong size %d expected %d", len(pc.data), 12)
	}

	_, err = pc.Read(&cp)

	if err != io.EOF {
		t.Errorf("Incorrect finalization")
	}

	buf2 := make([]byte, 10005)
	for i := range buf2 {
		buf2[i] = byte(i % 255)
	}

	binary.LittleEndian.PutUint32(buf2[1:], 10000)

	data, err = pc.Read(bytes.NewReader(buf2))
	if err != nil {
		t.Errorf("Error on reading large packet %v", err)
	} else {
		if len(data) != 10005 {
			t.Errorf("Wrong res data len %d", len(data))
		} else {
			for i, v := range data {
				if i >= 5 {
					if v != byte(i%255) {
						t.Errorf("Wrong byte %v expected %v on %d", v, (i % 255), i)
					}
				}
			}
		}
	}
}
*/

func Test_packetCombinerOverflow(t *testing.T) {
	cp := OneByteProvider{data: []byte{OP_EDONKEYHEADER, 0x04, 0xFF, 0x0F, 0xAB, OP_LOGINREQUEST, 0x01, 0x02, 0x03}}
	pc := PacketCombiner{data: make([]byte, 6)}
	_, _, err := pc.Read(&cp)

	if err == nil {
		t.Error("Overflow error expected")
	}
}

func Test_bufferCombiner(t *testing.T) {
	var bp = BufferProvider{bufs: [][]byte{{0x01, 0x04}, {0x00, 0x00, 0x00}, {0x10}}}
	buffer := make([]byte, 6)
	_, err := io.ReadFull(&bp, buffer)
	if err != nil {
		t.Errorf("Reading header buffer error %v", err)
	} else if !bytes.Equal(buffer, []byte{0x01, 0x04, 0x00, 0x00, 0x00, 0x10}) {
		t.Errorf("Received wrong bytes %v", buffer)
	}
}

func Test_usualPacket(t *testing.T) {
	var version uint32 = 0x3c
	var versionClient uint32 = 0x01
	var capability uint32 = 0x77

	var hello UsualPacket
	hello.Hash = LIBED2K
	hello.Point = Endpoint{Ip: 0, Port: 20033}
	hello.Properties = append(hello.Properties, CreateTag(version, CT_VERSION, ""))
	hello.Properties = append(hello.Properties, CreateTag(capability, CT_SERVER_FLAGS, ""))
	hello.Properties = append(hello.Properties, CreateTag(versionClient, CT_EMULE_VERSION, ""))
	hello.Properties = append(hello.Properties, CreateTag("ged2k", CT_NAME, ""))

	if len(hello.Properties) != 4 {
		t.Errorf("hello properties length incorrect %d", len(hello.Properties))
	}

	buf := make([]byte, 100)
	sb := StateBuffer{Data: buf}
	sb.Write(hello)
	wroteBytes := sb.Offset()
	if sb.err != nil {
		t.Errorf("Can not write hello %v", sb.Error())
	} else {
		sb2 := StateBuffer{Data: buf}
		var hello2 UsualPacket
		sb2.Read(&hello2)
		if sb2.Error() != nil {
			t.Errorf("Can not read hello %v", sb2.Error())
		} else {
			if len(hello2.Properties) != 4 {
				t.Errorf("Hello 2 prop len wrong %d", len(hello2.Properties))
			}

			if !bytes.Equal(hello.Hash[:], LIBED2K[:]) {
				t.Errorf("Hello 2 hash incorrect %x", hello2.Hash)
			}
		}
	}

	l := DataSize(hello)
	if l != wroteBytes {
		t.Errorf("Usual packet length %d does not match bytes have written %d", l, wroteBytes)
	}
}

func Test_ByteContainer(t *testing.T) {
	data := []byte{0x03, 0x00, 0x31, 0x32, 0x33}
	bc := ByteContainer{}
	sb := StateBuffer{Data: data}
	bc.Get(&sb)
	if sb.err != nil {
		t.Errorf("Get byte container from data error %v", sb.err)
	} else {
		if string(bc) != "123" {
			t.Errorf("As string error %x", bc)
		}
		recv := make([]byte, 5)
		sb2 := StateBuffer{Data: recv}
		bc.Put(&sb2)
		if sb2.err != nil {
			t.Errorf("Put byte container error %v", sb2.err)
		} else {
			if !bytes.Equal(recv, data) {
				t.Errorf("Wrote bytes %x do not match original %x", recv, data)
			}
		}
	}
}

func Test_IdSoftReading(t *testing.T) {
	data := []byte{0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00}
	id := IdChange{}
	sb := StateBuffer{Data: data}
	id.Get(&sb)
	if sb.err != nil {
		t.Errorf("Can not soft serialize %v", sb.err)
	} else {
		if id.ClientId != 1 || id.TcpFlags != 2 || id.AuxPort != 0 {
			t.Errorf("Wrong data on soft serialize")
		}
	}

	if sb.Remain() != 0 {
		t.Errorf("Remain incorrect %d expected 0", sb.Remain())
	}

	l := DataSize(id)
	if l != 12 {
		t.Errorf("Incorrect packet length %d", l)
	}
}

func Test_CollectionSize(t *testing.T) {
	data := []Endpoint{Endpoint{Ip: 1, Port: 2}, Endpoint{Ip: 3, Port: 4}}
	fs := FoundFileSources{Hash: EMULE, Sources: data}
	if DataSize(fs) != 16+1+12 {
		t.Errorf("Size of collection with two endpoint is wrong %d expected 16+1+12", DataSize(data))
	}
}

func Test_Endpoint2Str(t *testing.T) {
	template := []string{"0.0.0.22:1024",
		"0.0.0.16:3000",
		"127.0.0.1:2048",
		"196.127.10.1:30000",
		"192.168.0.33:50000",
		"255.255.255.255:60000",
		"88.122.32.45:10000"}

	ep1, _ := FromString("192.168.0.33:50000")
	ep2, _ := FromString("255.255.255.255:60000")
	ep3, _ := FromString("88.122.32.45:10000")

	endpoints := []Endpoint{Endpoint{Ip: 0x16000000, Port: 1024},
		Endpoint{Ip: 0x10000000, Port: 3000},
		Endpoint{Ip: 0x0100007f, Port: 2048},
		Endpoint{Ip: 0x010a7fc4, Port: 30000},
		ep1,
		ep2,
		ep3}

	for i := 0; i < len(template); i++ {
		if template[i] != endpoints[i].ToString() {
			t.Errorf("Endpoint to string %s does not match %s", endpoints[i].ToString(), template[i])
		}
	}

}

func Test_PacketCombiner(t *testing.T) {
	f, err := os.OpenFile("./search.dat", os.O_RDONLY, 0755)
	if err != nil {
		t.Errorf("Unable to open dat file %v", err)
	} else {
		pc := PacketCombiner{data: make([]byte, 100)}
		header, data, perr := pc.Read(f)
		if perr != nil {
			t.Errorf("Unable to process file %v", perr)
		} else {

			if header.Protocol != OP_PACKEDPROT {
				t.Errorf("Wrong protocol %v", header.Protocol)
			}

			sb := StateBuffer{Data: data}
			sr := SearchResult{}
			sb.Read(&sr)
			if sb.Error() != nil {
				t.Errorf("Unable to serialize search res %v", sb.Error())
			} else {
				if len(sr.Items) != 504 {
					t.Errorf("Wrong number of search res %d, expected 504", len(sr.Items))
				}

				if sr.MoreResults != 1 {
					t.Errorf("More results incorrect %v", sr.MoreResults)
				}
			}
		}
	}
}

func TestEndpoint_IsLocalAddress(t *testing.T) {
	endpoints := make([]Endpoint, 6)
	errs := make([]error, 6)
	expected := []bool{true, true, true, true, true, false}

	endpoints[0], errs[0] = FromString("10.0.0.3:4444")
	endpoints[1], errs[1] = FromString("172.16.22.33:5555")
	endpoints[2], errs[2] = FromString("192.168.1.1:5567")
	endpoints[3], errs[3] = FromString("169.254.2.2:1112")
	endpoints[4], errs[4] = FromString("127.1.1.45:5678")
	endpoints[5], errs[5] = FromString("217.115.10.22:2222")
	for i, x := range endpoints {
		if errs[i] != nil {
			t.Errorf("Failed to parse on index %d with %v", i, errs[i])
		} else if x.IsLocalAddress() != expected[i] {
			t.Errorf("Index %d failed", i)
		}
	}
}
