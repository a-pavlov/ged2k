package proto

type BitField struct {
	bytes []byte
	bits  int
}

func DivCeil(a int, b int) int {
	return (a + b - 1) / b
}

func BitsToBytes(count int) int {
	return DivCeil(count, 8)
}

func Min(a int, b int) int {
	if a < b {
		return a
	}

	return b
}

func (bf *BitField) Get(sb *StateBuffer) *StateBuffer {
	sz, err := sb.ReadUint16()
	if err == nil {
		bf.bits = int(sz)
		bf.bytes = make([]byte, BitsToBytes(int(sz)))
		sb.Read(bf.bytes)
	}

	return sb

}

func (bf BitField) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(uint16(len(bf.bytes))).Write(bf.bytes)
}

func (bf BitField) Size() int {
	return DataSize(uint16(1)) + len(bf.bytes)
}

func (bf *BitField) Resize(bits int) {
	b := BitsToBytes(bits)
	nbytes := make([]byte, b)
	copy(nbytes, bf.bytes[:Min(len(nbytes), len(bf.bytes))])
	bf.bytes = nbytes
	bf.bits = bits
	bf.ClearTrailingBits()
}

func (bf *BitField) ResizeVal(bits int, val bool) {
	s := bf.bits
	b := bf.bits & 7 // bits in last byte, reminder on size/8
	bf.Resize(bits)

	if s >= bf.bits {
		return
	}

	old_size_bytes := BitsToBytes(s)
	new_size_bytes := BitsToBytes(bf.bits)

	if val {
		if old_size_bytes != 0 && b != 0 {
			bf.bytes[old_size_bytes-1] |= (0xff >> b)
		}

		if old_size_bytes < new_size_bytes {
			for i := old_size_bytes; i < new_size_bytes; i++ {
				bf.bytes[i] = 0xff
			}
		}
		bf.ClearTrailingBits()
	} else {
		if old_size_bytes < new_size_bytes {
			for i := old_size_bytes; i < new_size_bytes; i++ {
				bf.bytes[i] = 0x00
			}
		}
	}
}

func (bf *BitField) ClearTrailingBits() {
	// clear the tail bits in the last byte
	if (bf.bits & 7) != 0 {
		bf.bytes[BitsToBytes(bf.bits)-1] &= 0xff << (8 - (bf.bits & 7))
	}
}

func (bf *BitField) SetAll() {
	for i := 0; i < len(bf.bytes); i++ {
		bf.bytes[i] = 0xff
	}

	bf.ClearTrailingBits()
}

func (bf *BitField) ClearAll() {
	for i := 0; i < len(bf.bytes); i++ {
		bf.bytes[i] = 0x00
	}
}
