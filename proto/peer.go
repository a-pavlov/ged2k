package proto

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
