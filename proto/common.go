package proto

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"reflect"
)

const MAX_ELEMS uint32 = 1000

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

func (sb *StateBuffer) ReadUint8() (uint8, error) {
	if sb.err != nil {
		return 0x0, sb.err
	}

	if sb.pos+1 < len(sb.Data) {
		res := sb.Data[sb.pos]
		sb.pos++
		return res, nil
	}

	return 0x0, io.EOF
}

func (sb *StateBuffer) ReadUint16() (uint16, error) {
	if sb.err != nil {
		return 0x0, sb.err
	}

	if sb.pos+2 <= len(sb.Data) {
		res := binary.LittleEndian.Uint16(sb.Data[sb.pos:])
		sb.pos = sb.pos + 2
		return res, nil
	}

	return 0x0, io.EOF
}

func (sb *StateBuffer) ReadUint32() (uint32, error) {
	if sb.err != nil {
		return 0x0, sb.err
	}

	if sb.pos+4 <= len(sb.Data) {
		res := binary.LittleEndian.Uint32(sb.Data[sb.pos:])
		sb.pos = sb.pos + 4
		return res, nil
	}

	return 0x0, io.EOF
}

func (sb *StateBuffer) ReadUint64() (uint64, error) {
	if sb.err != nil {
		return 0x0, sb.err
	}

	if sb.pos+8 <= len(sb.Data) {
		res := binary.LittleEndian.Uint64(sb.Data[sb.pos:])
		sb.pos = sb.pos + 8
		return res, nil
	}

	return 0x0, io.EOF
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

type Hash [16]byte

var Terminal = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x6A, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x89, 0xC0}
var LIBED2K = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x4C, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x4B, 0xC0}
var EMULE = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x0E, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x6F, 0xC0}

func (h *Hash) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(h[:])
}

func (h Hash) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(h[:])
}

func (h Hash) Size() int {
	return 16
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
	sz, err := sb.ReadUint32()
	if err == nil {
		if sz > MAX_ELEMS {
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
	H          Hash
	Point      Endpoint
	Properties TagCollection
}

func (up *UsualPacket) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&up.H).Read(&up.Point).Read(&up.Properties)
}

func (up UsualPacket) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(up.H).Write(up.Point).Write(up.Properties)
}

func (up UsualPacket) Size() int {
	return DataSize(up.H) + DataSize(up.Point) + DataSize(up.Properties)
}

type ByteContainer []byte

func (bc *ByteContainer) Get(sb *StateBuffer) *StateBuffer {
	var len uint16
	sb.Read(&len)
	if sb.err == nil {
		data := make([]byte, int(len))
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

	data, err := ReadTo(reader, pc.data[:6])

	if err != nil {
		return ph, data, err
	}

	ph.Read(data)

	fmt.Printf("Packet header %x/%d/%x\n", ph.Protocol, ph.Bytes, ph.Packet)

	if ph.Bytes > ED2K_MAX_PACKET_SIZE {
		return PacketHeader{}, data, fmt.Errorf("max packet size overflow %d", ph.Bytes)
	}

	if int(ph.Bytes) > len(pc.data) {
		// reallocate
		newSize := len(pc.data) * 2
		if int(ph.Bytes) > newSize {
			newSize = int(ph.Bytes)
		}

		fmt.Println("reallocate", newSize)

		buf := make([]byte, newSize)
		pc.data = buf
	}

	if ph.Bytes > 1 {
		_, err = ReadTo(reader, pc.data[:ph.Bytes-1])
		if err != nil {
			return ph, data, err
		}
	}

	if ph.Protocol == OP_PACKEDPROT {
		b := bytes.NewReader(pc.data[:ph.Bytes-1])
		z, err := zlib.NewReader(b)
		if err != nil {
			return ph, data, err
		}
		defer z.Close()
		unzipped, err := ioutil.ReadAll(z)
		if err != nil {
			return ph, data, err
		}

		// correct package size
		ph.Bytes = uint32(len(unzipped))
		return ph, unzipped, nil
	}

	return ph, pc.data[:ph.Bytes-1], nil
}

func ReadTo(r io.Reader, buffer []byte) ([]byte, error) {
	pos := 0
	for {
		n, err := r.Read(buffer[pos:])
		if err != nil {
			return buffer, err
		}

		pos += n

		if pos == len(buffer) {
			return buffer, nil
		}
	}
}
