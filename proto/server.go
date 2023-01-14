package proto

const SRV_TCPFLG_COMPRESSION = 0x00000001
const SRV_TCPFLG_NEWTAGS = 0x00000008
const SRV_TCPFLG_UNICODE = 0x00000010
const SRV_TCPFLG_RELATEDSEARCH = 0x00000040
const SRV_TCPFLG_TYPETAGINTEGER = 0x00000080
const SRV_TCPFLG_LARGEFILES = 0x00000100
const SRV_TCPFLG_TCPOBFUSCATION = 0x00000400

const SRVCAP_ZLIB = 0x0001
const SRVCAP_IP_IN_LOGIN = 0x0002
const SRVCAP_AUXPORT = 0x0004
const SRVCAP_NEWTAGS = 0x0008
const SRVCAP_UNICODE = 0x0010
const SRVCAP_LARGEFILES = 0x0100
const SRVCAP_SUPPORTCRYPT = 0x0200
const SRVCAP_REQUESTCRYPT = 0x0400
const SRVCAP_REQUIRECRYPT = 0x0800

const CAPABLE_ZLIB = SRVCAP_ZLIB
const CAPABLE_IP_IN_LOGIN_FRAME = SRVCAP_IP_IN_LOGIN
const CAPABLE_AUXPORT = SRVCAP_AUXPORT
const CAPABLE_NEWTAGS = SRVCAP_NEWTAGS
const CAPABLE_UNICODE = SRVCAP_UNICODE
const CAPABLE_LARGEFILES = SRVCAP_LARGEFILES

const GED2K_VERSION_MAJOR = 1
const GED2K_VERSION_MINOR = 1
const GED2K_VERSION_TINY = 0

type FoundFileSources struct {
	H       Hash
	Sources []Endpoint
}

func (fs *FoundFileSources) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&fs.H)
	sz, e := sb.ReadUint8()
	if e == nil {
		for i := 0; i < int(sz); i++ {
			ep := Endpoint{}
			sb.Read(&ep)
			fs.Sources = append(fs.Sources, ep)
			if sb.Error() != nil {
				break
			}
		}
	}

	return sb
}

func (fs *FoundFileSources) Put(sb *StateBuffer) *StateBuffer {
	var sz uint8 = uint8(len(fs.Sources))
	return sb.Write(fs.H).Write(sz).Write(fs.Sources)
}

func (fs FoundFileSources) Size() int {
	res := DataSize(byte(len(fs.Sources))) + DataSize(fs.H)
	for _, x := range fs.Sources {
		res += DataSize(x)
	}
	return res
}

type LoginRequest UsualPacket

type IdChange struct {
	ClientId uint32
	TcpFlags uint32
	AuxPort  uint32
}

func (i *IdChange) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&i.ClientId)
	if sb.err == nil && sb.Remain() >= 4 {
		sb.Read(&i.TcpFlags)
		if sb.err == nil && sb.Remain() >= 4 {
			sb.Read(&i.AuxPort)
		}
	}

	return sb
}

func (i IdChange) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(i.ClientId).Write(i.TcpFlags).Write(i.AuxPort)
}

func (i IdChange) Size() int {
	return DataSize(i.AuxPort) + DataSize(i.ClientId) + DataSize(i.TcpFlags)
}

type Status struct {
	UsersCount uint32
	FilesCount uint32
}

func (s *Status) Get(sb *StateBuffer) *StateBuffer {
	return sb.Read(&s.UsersCount).Read(&s.FilesCount)
}

func (s Status) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(s.UsersCount).Write(s.UsersCount)
}

func (s Status) Size() int {
	return DataSize(s.UsersCount) + DataSize(s.FilesCount)
}

type GetServerList struct{}

func (gl *GetServerList) Get(sb *StateBuffer) *StateBuffer {
	return sb
}

func (gl GetServerList) Put(sb *StateBuffer) *StateBuffer {
	return sb
}

func (gl GetServerList) Size() int {
	return 0
}
