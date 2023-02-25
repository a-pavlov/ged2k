package proto

import "fmt"

const PARTS_IN_REQUEST int = 3

const LARGE_FILE_OFFSET int = 4
const MULTIP_OFFSET int = 5
const SRC_EXT_OFFSET int = 10
const CAPTHA_OFFSET int = 11

type HelloAnswer struct {
	Hash        EMuleHash
	Point       Endpoint
	Properties  TagCollection
	ServerPoint Endpoint
}

func (ha *HelloAnswer) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&ha.Hash).Read(&ha.Point).Read(&ha.Properties).Read(&ha.ServerPoint)
}

func (ha HelloAnswer) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(ha.Hash).Write(ha.Point).Write(ha.Properties).Write(ha.ServerPoint)
}

func (ha HelloAnswer) Size() int {
	return DataSize(ha.Hash) + DataSize(ha.Point) + DataSize(ha.Properties) + DataSize(ha.ServerPoint)
}

type Hello struct {
	HashLength byte
	Answer     HelloAnswer
}

func (hello *Hello) Get(sb *StateBuffer) *StateBuffer {
	sb.ReadUint8()
	return sb.Read(hello.Answer)
}

func (hello Hello) Put(sb *StateBuffer) *StateBuffer {
	hlen := byte(0x10)
	return sb.Write(hlen).Write(hello.Answer)
}

func (hello Hello) Size() int {
	return DataSize(hello.HashLength) + DataSize(hello.Answer)
}

type ExtHello struct {
	Version         byte
	ProtocolVersion byte
	Properties      TagCollection
}

func (eh *ExtHello) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&eh.Version).Read(&eh.ProtocolVersion).Read(&eh.Properties)
}

func (eh ExtHello) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(eh.Version).Write(eh.ProtocolVersion).Write(eh.Properties)
}

func (eh ExtHello) Size() int {
	return DataSize(eh.Version) + DataSize(eh.ProtocolVersion) + DataSize(eh.Properties)
}

type FileAnswer struct {
	Hash EMuleHash
	Name ByteContainer
}

func (fa *FileAnswer) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&fa.Hash).Read(&fa.Name)
}

func (fa FileAnswer) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(fa.Hash).Write(fa.Name)
}

func (fa FileAnswer) Size() int {
	return DataSize(fa.Hash) + DataSize(fa.Name)
}

type FileStatusAnswer struct {
	Hash EMuleHash
	BF   BitField
}

func (fs *FileStatusAnswer) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&fs.Hash).Read(&fs.BF)
}

func (fs FileStatusAnswer) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(fs.Hash).Write(fs.BF)
}

func (fs FileStatusAnswer) Size() int {
	return DataSize(fs.Hash) + DataSize(fs.BF)
}

type HashSet struct {
	Hash        EMuleHash
	PieceHashes []EMuleHash
}

func (hs *HashSet) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&hs.Hash)
	size := sb.ReadUint16()
	if int(size) > MAX_ELEMS {
		sb.err = fmt.Errorf("elems count too large %v", size)
		return sb
	}

	hs.PieceHashes = make([]EMuleHash, int(size))
	for i, _ := range hs.PieceHashes {
		sb.Read(&hs.PieceHashes[i])
	}

	return sb
}

func (hs HashSet) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(hs.Hash).Write(uint16(len(hs.PieceHashes)))
	for _, x := range hs.PieceHashes {
		sb.Write(x)
	}
	return sb
}

func (hs HashSet) Size() int {
	return DataSize(hs.Hash) + DataSize(uint16(len(hs.PieceHashes))) + len(hs.PieceHashes)*DataSize(hs.Hash)
}

type RequestParts32 struct {
	Hash        EMuleHash
	BeginOffset [PARTS_IN_REQUEST]uint32
	EndOffset   [PARTS_IN_REQUEST]uint32
}

func (rp *RequestParts32) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&rp.Hash)
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		rp.BeginOffset[i] = sb.ReadUint32()
	}
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		rp.EndOffset[i] = sb.ReadUint32()
	}
	return sb
}

func (rp RequestParts32) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(rp.Hash)
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		sb.Write(rp.BeginOffset[i])
	}
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		sb.Write(rp.EndOffset[i])
	}
	return sb
}

func (rp RequestParts32) Size() int {
	return DataSize(rp.Hash) + DataSize(rp.BeginOffset[:])*2
}

type RequestParts64 struct {
	Hash        EMuleHash
	BeginOffset [PARTS_IN_REQUEST]uint64
	EndOffset   [PARTS_IN_REQUEST]uint64
}

func (rp *RequestParts64) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&rp.Hash)
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		rp.BeginOffset[i] = sb.ReadUint64()
	}
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		rp.EndOffset[i] = sb.ReadUint64()
	}
	return sb
}

func (rp RequestParts64) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(rp.Hash)
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		sb.Write(rp.BeginOffset[i])
	}
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		sb.Write(rp.EndOffset[i])
	}
	return sb
}

func (rp RequestParts64) Size() int {
	return DataSize(rp.Hash) + DataSize(rp.BeginOffset[:])*2
}

type MiscOptions struct {
	AichVersion         uint32
	UnicodeSupport      uint32
	UdpVer              uint32
	DataCompVer         uint32
	SupportSecIdent     uint32
	SourceExchange1Ver  uint32
	ExtendedRequestsVer uint32
	AcceptCommentVer    uint32
	NoViewSharedFiles   uint32
	MultiPacket         uint32
	SupportsPreview     uint32
}

func (mo MiscOptions) AsUint32() uint32 {
	return (mo.AichVersion << ((4 * 7) + 1)) |
		(mo.UnicodeSupport << 4 * 7) |
		(mo.UdpVer << 4 * 6) |
		(mo.DataCompVer << 4 * 5) |
		(mo.SupportSecIdent << 4 * 4) |
		(mo.SourceExchange1Ver << 4 * 3) |
		(mo.ExtendedRequestsVer << 4 * 2) |
		(mo.AcceptCommentVer << 4) |
		(mo.NoViewSharedFiles << 2) |
		(mo.MultiPacket << 1) |
		mo.SupportsPreview
}

func (mo *MiscOptions) Assign(value uint32) {
	mo.AichVersion = (value >> (4*7 + 1)) & 0x07
	mo.UnicodeSupport = (value >> 4 * 7) & 0x01
	mo.UdpVer = (value >> 4 * 6) & 0x0f
	mo.DataCompVer = (value >> 4 * 5) & 0x0f
	mo.SupportSecIdent = (value >> 4 * 4) & 0x0f
	mo.SourceExchange1Ver = (value >> 4 * 3) & 0x0f
	mo.ExtendedRequestsVer = (value >> 4 * 2) & 0x0f
	mo.AcceptCommentVer = (value >> 4) & 0x0f
	mo.NoViewSharedFiles = (value >> 2) & 0x01
	mo.MultiPacket = (value >> 1) & 0x01
	mo.SupportsPreview = value & 0x01
}

type MiscOptions2 uint32

func (mo MiscOptions2) SupportCaptcha() bool {
	return ((mo >> CAPTHA_OFFSET) & 0x01) == 1
}

func (mo MiscOptions2) SupportSourceExt2() bool {
	return ((mo >> SRC_EXT_OFFSET) & 0x01) == 1
}

func (mo MiscOptions2) SupportExtMultipacket() bool {
	return ((mo >> MULTIP_OFFSET) & 0x01) == 1
}

func (mo MiscOptions2) SupportLargeFiles() bool {
	return ((mo >> LARGE_FILE_OFFSET) & 0x01) == 0
}

func (mo *MiscOptions2) SetCaptcha() {
	*mo |= 1 << CAPTHA_OFFSET
}

func (mo *MiscOptions2) SetSourceExt2() {
	*mo |= 1 << SRC_EXT_OFFSET
}

func (mo *MiscOptions2) SetExtMultipacket() {
	*mo |= 1 << MULTIP_OFFSET
}

func (mo *MiscOptions2) SetLargeFiles() {
	*mo |= 1 << LARGE_FILE_OFFSET
}

type SendingPart struct {
	Hash     EMuleHash
	Begin    uint64
	End      uint64
	Extended bool
}

func (sp *SendingPart) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&sp.Hash)
	if sp.Extended {
		return sb.Read(&sp.Begin).Read(&sp.End)
	}

	sp.Begin = uint64(sb.ReadUint32())
	sp.End = uint64(sb.ReadUint32())
	return sb
}

func (sp *SendingPart) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(sp.Hash)
	if sp.Extended {
		return sb.Write(sp.Begin).Write(sp.End)
	}

	begin := uint32(sp.Begin)
	end := uint32(sp.End)
	return sb.Write(begin).Write(end)
}

func (sp SendingPart) Size() int {
	size := DataSize(sp.Begin)
	if sp.Extended {
		size *= 2
	}

	return size + DataSize(sp.Hash)
}

type CompressedPart struct {
	Hash                 EMuleHash
	Offset               uint64
	CompressedDataLength uint32
	Extended             bool
}

func (cp *CompressedPart) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&cp.Hash)
	if !cp.Extended {
		cp.Offset = uint64(sb.ReadUint32())
	} else {
		sb.Read(&cp.Offset)
	}

	return sb.Read(&cp.CompressedDataLength)
}

func (cp *CompressedPart) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(cp.Hash).Write(uint32(cp.Offset)).Write(cp.CompressedDataLength)
}

func (cp CompressedPart) Size() int {
	size := DataSize(uint32(0))
	if cp.Extended {
		size = DataSize(cp.Offset)
	}
	return size + DataSize(cp.Hash) + DataSize(cp.CompressedDataLength)
}
