package proto

import (
	"bytes"
	"encoding/binary"
	"io"
)

type UINT32 uint32
type UINT16 uint16
type UINT8 uint8
type BYTE byte

type Serializable interface {
	Get(sr *StateReader) *StateReader
	Put(sr *StateWriter) *StateWriter
}

type SerializableNumber interface {
	Get(sr *StateReader) *StateReader
	Put(sr *StateWriter) *StateWriter
	AsInteger() uint32
}

type StateReader struct {
	reader *bytes.Reader
	err    error
}

type StateWriter struct {
	buffer []byte
	err    error
	pos    int
}

func (sr *StateReader) read(data interface{}) *StateReader {
	if sr.err == nil {
		sr.err = binary.Read(sr.reader, binary.BigEndian, data)
	}
	return sr
}

// intDataSize returns the size of the data required to represent the data when encoded.
// It returns zero if the type cannot be implemented by the fast path in Read or Write.
func intDataSize(data interface{}) int {
	switch data := data.(type) {
	case bool, int8, uint8, *bool, *int8, *uint8, BYTE, UINT8:
		return 1
	case []bool:
		return len(data)
	case []int8:
		return len(data)
	case []uint8:
		return len(data)
	case int16, uint16, *int16, *uint16, UINT16:
		return 2
	case []int16:
		return 2 * len(data)
	case []uint16:
		return 2 * len(data)
	case int32, uint32, *int32, *uint32, UINT32:
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
	}
	return 0
}

func (sw *StateWriter) write(data interface{}) *StateWriter {
	if sw.err != nil {
		return sw
	}

	n := intDataSize(data)
	if sw.pos+n > len(sw.buffer) {
		sw.err = io.EOF
		return sw
	}

	switch v := data.(type) {
	case int8:
		sw.buffer[0] = byte(v)
	case uint16:
		binary.BigEndian.PutUint16(sw.buffer, v)
	case uint32:
		binary.BigEndian.PutUint32(sw.buffer, v)
	case UINT32:
		binary.BigEndian.PutUint32(sw.buffer, uint32(v))
	case UINT16:
		binary.BigEndian.PutUint16(sw.buffer, uint16(v))
	case *[]byte:
		for i, x := range *v {
			sw.buffer[i] = byte(x)
		}
	case *Hash:
		for i, x := range v {
			sw.buffer[sw.pos+i] = byte(x)
		}
	default:
		panic("No serializable")
	}

	sw.pos += n
	return sw
}

func (sw StateWriter) Bytes() int {
	return sw.pos
}

func (sw StateWriter) Error() error {
	return sw.err
}

func (s *UINT32) Get(sr *StateReader) *StateReader {
	return sr.read(s)
}

func (s *UINT32) Put(sw *StateWriter) *StateWriter {
	return sw.write(*s)
}

func (s *UINT16) Get(sr *StateReader) *StateReader {
	return sr.read(s)
}

func (s *UINT16) Put(sw *StateWriter) *StateWriter {
	return sw.write(*s)
}

func (s *BYTE) Get(sr *StateReader) *StateReader {
	return sr.read(s)
}

func (s *BYTE) Put(sw *StateWriter) *StateWriter {
	if sw.err != nil {
		return sw
	}

	if sw.pos+1 > len(sw.buffer) {
		sw.err = io.EOF
	} else {
		sw.buffer[sw.pos] = byte(*s)
		sw.pos += 1
	}

	return sw
}

type Some struct {
	Ip   UINT16
	Port UINT16
}

func (s *Some) Get(sr *StateReader) *StateReader {
	return s.Port.Get(s.Ip.Get(sr))
}

func (s *Some) Put(sw *StateWriter) *StateWriter {
	return s.Port.Put(s.Ip.Put(sw))
}

type Hash [16]byte

var Terminal = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x6A, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x89, 0xC0}
var LIBED2K = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x4C, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x4B, 0xC0}
var EMULE = [16]byte{0x31, 0xD6, 0xCF, 0xE0, 0xD1, 0x0E, 0xE9, 0x31, 0xB7, 0x3C, 0x59, 0xD7, 0xE0, 0xC0, 0x6F, 0xC0}

func (h *Hash) Get(sr *StateReader) *StateReader {
	return sr.read(h)
}

func (h *Hash) Put(sw *StateWriter) *StateWriter {
	return sw.write(h)
}
