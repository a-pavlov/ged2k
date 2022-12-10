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

const GED2K_VERSION_MAJOR = 0
const GED2K_VERSION_MINOR = 1
const GED2K_VERSION_TINY = 0

type FoundFileSources struct {
	H       Hash
	Sources Collection
}

func (fs *FoundFileSources) Get(sb *StateBuffer) *StateBuffer {
	sb.Read(&fs.H)
	sz, e := sb.ReadUint8()
	if e == nil {
		for i := 0; i < int(sz); i++ {
			fs.Sources = append(fs.Sources, &Endpoint{})
		}

		sb.Read(&fs.Sources)
	}

	return sb
}

func (fs *FoundFileSources) Put(sb *StateBuffer) *StateBuffer {
	var sz uint8 = uint8(len(fs.Sources))
	return sb.Write(fs.H).Write(sz).Write(fs.Sources)
}

type LoginRequest UsualPacket
