package proto

type BitField struct {
	bytes []byte
	bits  int
}

func DivCeilInt(a int, b int) int {
	return (a + b - 1) / b
}

func BitsToBytes(count int) int {
	return DivCeilInt(count, 8)
}

func Min(a int, b int) int {
	if a < b {
		return a
	}

	return b
}

func (bf *BitField) Get(sb *StateBuffer) *StateBuffer {
	sz := sb.ReadUint16()
	if sb.Error() == nil {
		bf.bits = int(sz)
		if bf.bits > 0 {
			bf.bytes = make([]byte, BitsToBytes(int(sz)))
			sb.Read(bf.bytes)
		}
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

func (bf *BitField) Assign(b []byte, c int) {
	bf.Resize(c)
	copy(bf.bytes, b[:BitsToBytes(c)])
	bf.ClearTrailingBits()
}

func (bf BitField) GetBit(index int) bool {
	return (bf.bytes[index/8] & (0x80 >> (index & 7))) != 0
}

func (bf *BitField) ClearBit(index int) {
	bf.bytes[index/8] &= ^(0x80 >> (index & 7))
}

func (bf *BitField) SetBit(index int) {
	bf.bytes[index/8] |= (0x80 >> (index & 7))
}

func (bf BitField) Bits() int {
	return bf.bits
}

func (bf BitField) Count() int {
	// 0000, 0001, 0010, 0011, 0100, 0101, 0110, 0111,
	// 1000, 1001, 1010, 1011, 1100, 1101, 1110, 1111
	num_bits := []byte{
		0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}

	ret := 0
	num_bytes := bf.bits / 8
	for i := 0; i < num_bytes; i++ {
		ret += int(num_bits[bf.bytes[i]&0xf]) + int(num_bits[bf.bytes[i]>>4])
	}

	rest := bf.bits - num_bytes*8
	for i := 0; i < rest; i++ {
		ret += int((bf.bytes[num_bytes] >> (7 - i)) & 1)
	}

	return ret
}
