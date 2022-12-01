package proto

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
)

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

/*func (sr *StateReader) read(data interface{}) *StateReader {
	if sr.err == nil {

		sr.err = binary.Read(sr.reader, binary.LittleEndian, data)
	}
	return sr
}*/

func (sb *StateBuffer) read(data interface{}) *StateBuffer {
	if sb.err != nil {
		return sb
	}

	// Fast path for basic types and slices.
	if n := intDataSize(data); n != 0 {
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
		case *Hash:
			for i, x := range bs {
				data[i] = x
			}
		case *ByteContainer16:
			for i, x := range bs {
				([]byte(*data))[i] = x
			}
		case *ByteContainer32:
			for i, x := range bs {
				([]byte(*data))[i] = x
			}
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
			n = 0 // fast path doesn't apply
		}

		if n != 0 {
			sb.pos += n
		} else {
			sb.err = errors.New("SB.Read: invalid type " + reflect.TypeOf(data).String())
		}
	} else {
		sb.err = errors.New("SB.Read: invalid type " + reflect.TypeOf(data).String())
	}

	return sb

	// Fallback to reflect-based decoding.
	/*
		v := reflect.ValueOf(data)
		size := -1
		switch v.Kind() {
		case reflect.Ptr:
			v = v.Elem()
			size = dataSize(v)
		case reflect.Slice:
			size = dataSize(v)
		}
		if size < 0 {
			return errors.New("binary.Read: invalid type " + reflect.TypeOf(data).String())
		}
		d := &decoder{order: order, buf: make([]byte, size)}
		if _, err := io.ReadFull(r, d.buf); err != nil {
			return err
		}
		d.value(v)
		return nil*/
}

// intDataSize returns the size of the data required to represent the data when encoded.
// It returns zero if the type cannot be implemented by the fast path in Read or Write.
func intDataSize(data interface{}) int {
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
	case Hash:
		return 16
	case *Hash:
		return 16
	case ByteContainer32:
		return len(data)
	case *ByteContainer32:
		return len(*data)
	case ByteContainer16:
		return len(data)
	case *ByteContainer16:
		return len(*data)
	}
	return 0
}

func (sb *StateBuffer) write(data interface{}) *StateBuffer {
	if sb.err != nil {
		return sb
	}

	n := intDataSize(data)
	if sb.pos+n > len(sb.Data) {
		sb.err = io.EOF
		return sb
	}

	switch v := data.(type) {
	case int8:
		sb.Data[0] = byte(v)
	case uint16:
		binary.LittleEndian.PutUint16(sb.Data, v)
	case uint32:
		binary.LittleEndian.PutUint32(sb.Data, v)
	case *[]byte:
		for i, x := range *v {
			sb.Data[sb.pos+i] = byte(x)
		}
	case []byte:
		for i, x := range v {
			sb.Data[sb.pos+i] = byte(x)
		}
	case *Hash:
		for i, x := range v {
			sb.Data[sb.pos+i] = byte(x)
		}
	case ByteContainer32:
		for i, x := range v {
			sb.Data[sb.pos+i] = byte(x)
		}
	case *ByteContainer32:
		for i, x := range *v {
			sb.Data[sb.pos+i] = byte(x)
		}
	case ByteContainer16:
		for i, x := range v {
			sb.Data[sb.pos+i] = byte(x)
		}
	case *ByteContainer16:
		for i, x := range *v {
			sb.Data[sb.pos+i] = byte(x)
		}
	default:
		sb.err = errors.New("SB.write: invalid type " + reflect.TypeOf(data).String())
		return sb
	}

	sb.pos += n
	return sb
}

func (sw StateBuffer) Bytes() int {
	return sw.pos
}

func (sw StateBuffer) Error() error {
	return sw.err
}

type Some struct {
	Ip   uint16
	Port uint16
}

func (s *Some) Get(sb *StateBuffer) *StateBuffer {
	return sb.read(s.Ip).read(s.Port)
}

func (s *Some) Put(sb *StateBuffer) *StateBuffer {
	return sb.write(s.Ip).write(s.Port)
}

type Hash [16]byte

var Terminal = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x6A, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x89, 0xC0}
var LIBED2K = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x4C, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x4B, 0xC0}
var EMULE = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x0E, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x6F, 0xC0}

func (h *Hash) Get(sb *StateBuffer) *StateBuffer {
	return sb.read(h)
}

func (h *Hash) Put(sb *StateBuffer) *StateBuffer {
	return sb.write(h)
}

type ByteContainer16 []byte

func (bc *ByteContainer16) Get(sb *StateBuffer) *StateBuffer {
	var size uint16
	sb.read(&size)
	if sb.err == nil {
		*bc = make([]byte, size)
		sb.read(bc)
	}

	return sb
}

func (bc *ByteContainer16) Put(sb *StateBuffer) *StateBuffer {
	var size uint16 = uint16(len(*bc))
	sb.write(size)
	if sb.err == nil {
		sb.write(bc)
	}

	return sb
}

type ByteContainer32 []byte

func (bc *ByteContainer32) Get(sb *StateBuffer) *StateBuffer {
	var size uint32
	sb.read(&size)
	if sb.err == nil {
		*bc = make([]byte, size)
		sb.read(bc)
	}

	return sb
}

func (bc *ByteContainer32) Put(sb *StateBuffer) *StateBuffer {
	var size uint32 = uint32(len(*bc))
	sb.write(size)
	if sb.err == nil {
		sb.write(bc)
	}

	return sb
}

type IP struct {
	ip   uint32
	port uint16
}

func (i *IP) Get(sb *StateBuffer) *StateBuffer {
	return sb.read(i.ip).read(i.port)
}

func (i *IP) Put(sb *StateBuffer) *StateBuffer {
	return sb.write(i.ip).write(i.port)
}
