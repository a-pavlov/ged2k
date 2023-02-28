package proto

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
)

const MAX_ELEMS int = 1000

type Serializable interface {
	Get(sr *StateBuffer) *StateBuffer
	Put(sr *StateBuffer) *StateBuffer
}

type SerializableNumber interface {
	Get(sr *StateBuffer) *StateBuffer
	Put(sr *StateBuffer) *StateBuffer
	AsInteger() uint32
}

type StateBuffer struct {
	Data []byte
	err  error
	pos  int
}

func (sb *StateBuffer) ReadUint8() uint8 {
	if sb.err != nil {
		return 0x0
	}

	if sb.pos+1 <= len(sb.Data) {
		res := sb.Data[sb.pos]
		sb.pos++
		return res
	}

	sb.err = io.EOF
	return 0x0
}

func (sb *StateBuffer) ReadUint16() uint16 {
	if sb.err != nil {
		return 0x0
	}

	if sb.pos+2 <= len(sb.Data) {
		res := binary.LittleEndian.Uint16(sb.Data[sb.pos:])
		sb.pos = sb.pos + 2
		return res
	}

	sb.err = io.EOF
	return 0x0
}

func (sb *StateBuffer) ReadUint32() uint32 {
	if sb.err != nil {
		return 0x0
	}

	if sb.pos+4 <= len(sb.Data) {
		res := binary.LittleEndian.Uint32(sb.Data[sb.pos:])
		sb.pos = sb.pos + 4
		return res
	}

	sb.err = io.EOF
	return 0x0
}

func (sb *StateBuffer) ReadUint64() uint64 {
	if sb.err != nil {
		return 0x0
	}

	if sb.pos+8 <= len(sb.Data) {
		res := binary.LittleEndian.Uint64(sb.Data[sb.pos:])
		sb.pos = sb.pos + 8
		return res
	}

	sb.err = io.EOF
	return 0x0
}

func (sb *StateBuffer) Read(data interface{}) *StateBuffer {
	if sb.err != nil {
		return sb
	}

	st, ok := data.(interface {
		Get(sb *StateBuffer) *StateBuffer
	})

	if ok {
		return st.Get(sb)
	}

	if n := DataSize(data); n >= 0 {
		if n+sb.pos > len(sb.Data) {
			sb.err = io.EOF
			return sb
		}

		bs := sb.Data[sb.pos : sb.pos+n]

		switch data := data.(type) {
		case *bool:
			*data = bs[0] != 0
		case *int8:
			*data = int8(bs[0])
		case *uint8:
			*data = bs[0]
		case *int16:
			*data = int16(binary.LittleEndian.Uint16(bs))
		case *uint16:
			*data = binary.LittleEndian.Uint16(bs)
		case *int32:
			*data = int32(binary.LittleEndian.Uint32(bs))
		case *uint32:
			*data = binary.LittleEndian.Uint32(bs)
		case *int64:
			*data = int64(binary.LittleEndian.Uint64(bs))
		case *uint64:
			*data = binary.LittleEndian.Uint64(bs)
		case *float32:
			*data = math.Float32frombits(binary.LittleEndian.Uint32(bs))
		case *float64:
			*data = math.Float64frombits(binary.LittleEndian.Uint64(bs))
		case []bool:
			for i, x := range bs { // Easier to loop over the input for 8-bit values.
				data[i] = x != 0
			}
		case []int8:
			for i, x := range bs {
				data[i] = int8(x)
			}
		case []uint8:
			copy(data, bs)
		case []int16:
			for i := range data {
				data[i] = int16(binary.LittleEndian.Uint16(bs[2*i:]))
			}
		case []uint16:
			for i := range data {
				data[i] = binary.LittleEndian.Uint16(bs[2*i:])
			}
		case []int32:
			for i := range data {
				data[i] = int32(binary.LittleEndian.Uint32(bs[4*i:]))
			}
		case []uint32:
			for i := range data {
				data[i] = binary.LittleEndian.Uint32(bs[4*i:])
			}
		case []int64:
			for i := range data {
				data[i] = int64(binary.LittleEndian.Uint64(bs[8*i:]))
			}
		case []uint64:
			for i := range data {
				data[i] = binary.LittleEndian.Uint64(bs[8*i:])
			}
		case []float32:
			for i := range data {
				data[i] = math.Float32frombits(binary.LittleEndian.Uint32(bs[4*i:]))
			}
		case []float64:
			for i := range data {
				data[i] = math.Float64frombits(binary.LittleEndian.Uint64(bs[8*i:]))
			}
		default:
			n = 0
		}

		if n != 0 {
			sb.pos += n
			return sb
		}
	}

	sb.err = errors.New("SB.Read: invalid type " + reflect.TypeOf(data).String())
	return sb
}

// DataSize returns the size of the data required to represent the data when encoded.
// It returns zero if the type cannot be implemented by the fast path in Read or Write.
func DataSize(data interface{}) int {

	st, ok := data.(interface {
		Size() int
	})

	if ok {
		return st.Size()
	}

	switch data := data.(type) {
	case bool, int8, uint8, *bool, *int8, *uint8:
		return 1
	case []bool:
		return len(data)
	case []int8:
		return len(data)
	case []uint8:
		return len(data)
	case int16, uint16, *int16, *uint16:
		return 2
	case []int16:
		return 2 * len(data)
	case []uint16:
		return 2 * len(data)
	case int32, uint32, *int32, *uint32:
		return 4
	case []int32:
		return 4 * len(data)
	case []uint32:
		return 4 * len(data)
	case int64, uint64, *int64, *uint64:
		return 8
	case []int64:
		return 8 * len(data)
	case []uint64:
		return 8 * len(data)
	case float32, *float32:
		return 4
	case float64, *float64:
		return 8
	case []float32:
		return 4 * len(data)
	case []float64:
		return 8 * len(data)
	}
	panic("Can not obtain size of type " + reflect.TypeOf(data).String())
}

func (sb *StateBuffer) Write(data interface{}) *StateBuffer {
	if sb.err != nil {
		return sb
	}

	st, ok := data.(interface {
		Put(sb *StateBuffer) *StateBuffer
	})
	if ok {
		return st.Put(sb)
	}

	n := DataSize(data)

	if sb.pos+n > len(sb.Data) {
		sb.err = io.EOF
		return sb
	}

	switch v := data.(type) {
	case uint8:
		sb.Data[sb.pos] = byte(v)
	case uint16:
		binary.LittleEndian.PutUint16(sb.Data[sb.pos:], v)
	case uint32:
		binary.LittleEndian.PutUint32(sb.Data[sb.pos:], v)
	case uint64:
		binary.LittleEndian.PutUint64(sb.Data[sb.pos:], v)
	case *[]byte:
		for i, x := range *v {
			sb.Data[sb.pos+i] = byte(x)
		}
	case []byte:
		for i, x := range v {
			sb.Data[sb.pos+i] = byte(x)
		}
	case []interface{}:
		for _, x := range v {
			sb.Write(x)
		}
	default:
		sb.err = errors.New("SB.write: invalid type " + reflect.TypeOf(data).String())
		return sb
	}

	sb.pos += n
	return sb
}

func (sw StateBuffer) Offset() int {
	return sw.pos
}

func (sw StateBuffer) Error() error {
	return sw.err
}

func (sb StateBuffer) Remain() int {
	return len(sb.Data) - sb.pos
}

type Some struct {
	Ip   uint16
	Port uint16
}

func (s *Some) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(s.Ip).Read(s.Port)
}

func (s *Some) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(s.Ip).Write(s.Port)
}

const HASH_LEN int = 16

type ED2KHash [HASH_LEN]byte

var Terminal = ED2KHash{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x6A, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x89, 0xC0}
var LIBED2K = ED2KHash{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x4C, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x4B, 0xC0}
var EMULE = ED2KHash{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x0E, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x6F, 0xC0}
var ZERO = ED2KHash{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

func (h *ED2KHash) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(h[:])
}

func (h ED2KHash) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(h[:])
}

func (h ED2KHash) Size() int {
	return 16
}

func (h ED2KHash) Equals(hash ED2KHash) bool {
	return bytes.Equal(h[:], hash[:])
}

func (h ED2KHash) ToString() string {
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

func String2Hash(s string) ED2KHash {
	var h ED2KHash
	src := []byte(s)
	n, err := hex.Decode(h[:], src)
	if err != nil || n != HASH_LEN {
		return ZERO
	}

	return h
}

type Endpoint struct {
	Ip   uint32
	Port uint16
}

func (i *Endpoint) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&i.Ip).Read(&i.Port)
}

func (i Endpoint) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(i.Ip).Write(i.Port)
}

func (i Endpoint) Size() int {
	return DataSize(i.Ip) + DataSize(i.Port)
}

type IP uint32

func (ip *IP) Get(sb *StateBuffer) *StateBuffer {
	*ip = IP(sb.ReadUint32())
	return sb
}

func (ip IP) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(uint32(ip))
}

func (ip IP) Size() int {
	return DataSize(uint32(ip))
}

func (i Endpoint) AsString() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", i.Ip&0xff, (i.Ip>>8)&0xff, (i.Ip>>16)&0xff, (i.Ip>>24)&0xff, i.Port)
}

func FromString(s string) (Endpoint, error) {
	parts := strings.Split(s, ":")

	if len(parts) != 2 {
		return Endpoint{}, fmt.Errorf("can not find ip-port separator: %s", s)
	}

	ips := strings.Split(parts[0], ".")

	if len(ips) != 4 {
		return Endpoint{}, fmt.Errorf("ip has no correct format")
	}

	i, err := strconv.Atoi(ips[0])
	i_1, err_1 := strconv.Atoi(ips[1])
	i_2, err_2 := strconv.Atoi(ips[2])
	i_3, err_3 := strconv.Atoi(ips[3])
	port, err_p := strconv.Atoi(parts[1])

	if err != nil || err_1 != nil || err_2 != nil || err_3 != nil || err_p != nil {
		return Endpoint{}, fmt.Errorf("no number found")
	}

	return Endpoint{Ip: uint32(i) | ((uint32(i_1) << 8) & 0xff00) | ((uint32(i_2) << 16) & 0xff0000) | ((uint32(i_3) << 24) & 0xff000000), Port: uint16(port)}, nil
}

func EndpointFromString(s string) Endpoint {
	p, e := FromString(s)
	if e != nil {
		return Endpoint{}
	}

	return p
}

func GetContainer(data []Serializable, sb *StateBuffer) {
	for _, x := range data {
		x.Get(sb)
		if sb.err != nil {
			break
		}
	}
}

type Collection []Serializable
type TagCollection []Tag

func (c *Collection) Get(sb *StateBuffer) *StateBuffer {
	for i := 0; i < len(*c); i++ {
		(*c)[i].Get(sb)
	}

	return sb
}

func (c Collection) Put(sb *StateBuffer) *StateBuffer {
	for i := 0; i < len(c); i++ {
		c[i].Put(sb)
	}

	return sb
}

func (c *TagCollection) Get(sb *StateBuffer) *StateBuffer {
	sz := sb.ReadUint32()
	if sb.Error() == nil {
		if int(sz) > MAX_ELEMS {
			sb.err = fmt.Errorf("elements count greater than max elements %d", sz)
			return sb
		}

		for i := 0; i < int(sz); i++ {
			t := Tag{}
			sb.Read(&t)
			*c = append(*c, t)
			if sb.err != nil {
				break
			}
		}
	}

	return sb
}

func (c TagCollection) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(uint32(len(c)))
	for i := 0; i < len(c); i++ {
		c[i].Put(sb)
	}

	return sb
}

func (c TagCollection) Size() int {
	res := DataSize(uint32(1))

	for i := 0; i < len(c); i++ {
		res += DataSize(c[i])
	}

	return res
}

type UsualPacket struct {
	Hash       ED2KHash
	Point      Endpoint
	Properties TagCollection
}

func (up *UsualPacket) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&up.Hash).Read(&up.Point).Read(&up.Properties)
}

func (up UsualPacket) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(up.Hash).Write(up.Point).Write(up.Properties)
}

func (up UsualPacket) Size() int {
	return DataSize(up.Hash) + DataSize(up.Point) + DataSize(up.Properties)
}

type ByteContainer []byte

func (bc *ByteContainer) Get(sb *StateBuffer) *StateBuffer {
	length := sb.ReadUint16()
	if sb.err == nil {
		data := make([]byte, int(length))
		sb.Read(data)
		if sb.err == nil {
			*bc = data
		}
	}

	return sb
}

func (bc ByteContainer) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(uint16(len(bc))).Write([]byte(bc))
}

func (bc ByteContainer) Size() int {
	return DataSize(uint16(1)) + DataSize([]byte(bc[:]))
}

func String2ByteContainer(s string) ByteContainer {
	return []byte(s)
}

func (bc ByteContainer) ToString() string {
	return string(bc)
}

type PacketHeader struct {
	Protocol byte
	Bytes    uint32
	Packet   byte
}

func (ph *PacketHeader) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&ph.Packet).Read(&ph.Bytes).Read(&ph.Packet)
}

func (ph PacketHeader) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(ph.Protocol).Write(ph.Bytes).Write(ph.Packet)
}

func (ph PacketHeader) IsEmpty() bool {
	return ph.Protocol == 0x0 && ph.Bytes == 0 && ph.Packet == 0x0
}

func (ph PacketHeader) Size() int {
	return DataSize(ph.Protocol) + DataSize(ph.Bytes) + DataSize(ph.Packet)
}

func (ph *PacketHeader) Reset() {
	ph.Bytes = 0
	ph.Packet = 0x0
	ph.Protocol = 0x0
}

func (ph *PacketHeader) Read(buffer []byte) {
	ph.Protocol = buffer[0]
	ph.Bytes = binary.LittleEndian.Uint32(buffer[1:])
	ph.Packet = buffer[5]
}

func (ph PacketHeader) Write(buffer []byte) {
	buffer[0] = ph.Protocol
	binary.LittleEndian.PutUint32(buffer[1:], ph.Bytes)
	buffer[5] = ph.Packet
}

type PacketCombiner struct {
	data []byte
}

func (pc *PacketCombiner) Read(reader io.Reader) (PacketHeader, []byte, error) {
	ph := PacketHeader{}

	if pc.data == nil {
		pc.data = make([]byte, 6)
	}

	_, err := io.ReadFull(reader, pc.data[:6])

	if err != nil {
		return ph, pc.data, err
	}

	ph.Read(pc.data[:6])

	log.Printf("Packet header HEX protocol/packet:[%x][%x] DEC bytes: %d\n", ph.Protocol, ph.Packet, ph.Bytes)

	bytesToRead := int(ph.Bytes)
	switch {
	case ph.Protocol == OP_EDONKEYPROT && ph.Packet == OP_SENDINGPART:
		bytesToRead = SendingPart{Extended: false}.Size()
	case ph.Protocol == OP_EMULEPROT && ph.Packet == OP_SENDINGPART_I64:
		bytesToRead = SendingPart{Extended: true}.Size()
	case ph.Protocol == OP_EMULEPROT && ph.Packet == OP_COMPRESSEDPART:
		bytesToRead = CompressedPart{Extended: false}.Size()
	case ph.Protocol == OP_EMULEPROT && ph.Packet == OP_COMPRESSEDPART_I64:
		bytesToRead = CompressedPart{Extended: true}.Size()
	default:
		bytesToRead = bytesToRead - 1
	}

	if bytesToRead > ED2K_MAX_PACKET_SIZE {
		return PacketHeader{}, pc.data[:6], fmt.Errorf("max packet size overflow %d", ph.Bytes)
	}

	if bytesToRead > len(pc.data) {
		// reallocate
		newSize := len(pc.data) * 2
		if bytesToRead > newSize {
			newSize = bytesToRead
		}

		log.Println("reallocate", newSize)

		buf := make([]byte, newSize)
		pc.data = buf
	}

	if bytesToRead > 0 {
		_, err = io.ReadFull(reader, pc.data[:bytesToRead])
		if err != nil {
			return ph, pc.data[:6], err
		}
	}

	if ph.Protocol == OP_PACKEDPROT {
		b := bytes.NewReader(pc.data[:bytesToRead])
		///err := os.WriteFile("/tmp/dat1", pc.data[:bytesToRead], 0644)
		z, err := zlib.NewReader(b)
		if err != nil {
			return ph, pc.data[:6], err
		}
		defer z.Close()
		unzipped, err := ioutil.ReadAll(z)
		if err != nil {
			return ph, pc.data[:6], err
		}

		// correct package size
		ph.Bytes = uint32(len(unzipped))
		return ph, unzipped, nil
	}

	return ph, pc.data[:bytesToRead], nil
}

func InBlockOffset(begin uint64) int {
	return int(begin % PIECE_SIZE_UINT64)
}

func DivCeil64(a uint64, b uint64) uint64 {
	return (a + b - 1) / b
}

func NumPiecesAndBlocks(offset uint64) (int, int) {
	if offset == 0 {
		return 0, 0
	}
	blocksInLastPiece := (int)(DivCeil64(offset%PIECE_SIZE_UINT64, BLOCK_SIZE_UINT64))
	if blocksInLastPiece == 0 {
		blocksInLastPiece = BLOCKS_PER_PIECE
	}
	return int(DivCeil64(offset, PIECE_SIZE_UINT64)), blocksInLastPiece
}
