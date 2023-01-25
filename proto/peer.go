package proto

const PARTS_IN_REQUEST int = 3

type HelloAnswer struct {
	H           Hash
	Point       Endpoint
	Properties  TagCollection
	ServerPoint Endpoint
}

func (ha *HelloAnswer) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&ha.H).Read(&ha.Point).Read(&ha.Properties).Read(&ha.ServerPoint)
}

func (ha HelloAnswer) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(ha.H).Write(ha.Point).Write(ha.Properties).Write(ha.ServerPoint)
}

func (ha HelloAnswer) Size() int {
	return DataSize(ha.H) + DataSize(ha.Point) + DataSize(ha.Properties) + DataSize(ha.ServerPoint)
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
	H    Hash
	Name ByteContainer
}

func (fa *FileAnswer) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&fa.H).Read(&fa.Name)
}

func (fa FileAnswer) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(fa.H).Write(fa.Name)
}

func (fa FileAnswer) Size() int {
	return DataSize(fa.H) + DataSize(fa.Name)
}

type FileStatusAnswer struct {
	H  Hash
	BF BitField
}

func (fs *FileStatusAnswer) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&fs.H).Read(&fs.BF)
}

func (fs FileStatusAnswer) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(fs.H).Write(fs.BF)
}

func (fs FileStatusAnswer) Size() int {
	return DataSize(fs.H) + DataSize(fs.BF)
}

type RequestParts32 struct {
	H           Hash
	BeginOffset [PARTS_IN_REQUEST]uint32
	EndOffset   [PARTS_IN_REQUEST]uint32
}

func (rp *RequestParts32) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&rp.H)
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		rp.BeginOffset[i], _ = sb.ReadUint32()
	}
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		rp.EndOffset[i], _ = sb.ReadUint32()
	}
	return sb
}

func (rp RequestParts32) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(rp.H)
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		sb.Write(rp.BeginOffset[i])
	}
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		sb.Write(rp.EndOffset[i])
	}
	return sb
}

func (rp RequestParts32) Size() int {
	return DataSize(rp.H) + DataSize(rp.BeginOffset[:])*2
}

type RequestParts64 struct {
	H           Hash
	BeginOffset [PARTS_IN_REQUEST]uint64
	EndOffset   [PARTS_IN_REQUEST]uint64
}

func (rp *RequestParts64) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&rp.H)
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		rp.BeginOffset[i], _ = sb.ReadUint64()
	}
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		rp.EndOffset[i], _ = sb.ReadUint64()
	}
	return sb
}

func (rp RequestParts64) Put(sb *StateBuffer) *StateBuffer {
	sb.Write(rp.H)
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		sb.Write(rp.BeginOffset[i])
	}
	for i := 0; i < PARTS_IN_REQUEST; i++ {
		sb.Write(rp.EndOffset[i])
	}
	return sb
}

func (rp RequestParts64) Size() int {
	return DataSize(rp.H) + DataSize(rp.BeginOffset[:])*2
}
