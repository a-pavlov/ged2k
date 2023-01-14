package proto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

const TAGTYPE_UNDEFINED byte = 0x00 // special tag definition for empty objects
const TAGTYPE_HASH16 byte = 0x01
const TAGTYPE_STRING byte = 0x02
const TAGTYPE_UINT32 byte = 0x03
const TAGTYPE_FLOAT32 byte = 0x04
const TAGTYPE_BOOL byte = 0x05
const TAGTYPE_BOOLARRAY byte = 0x06
const TAGTYPE_BLOB byte = 0x07
const TAGTYPE_UINT16 byte = 0x08
const TAGTYPE_UINT8 byte = 0x09
const TAGTYPE_BSOB byte = 0x0A
const TAGTYPE_UINT64 byte = 0x0B

// Compressed string types
const TAGTYPE_STR1 byte = 0x11
const TAGTYPE_STR2 byte = 0x12
const TAGTYPE_STR3 byte = 0x13
const TAGTYPE_STR4 byte = 0x14
const TAGTYPE_STR5 byte = 0x15
const TAGTYPE_STR6 byte = 0x16
const TAGTYPE_STR7 byte = 0x17
const TAGTYPE_STR8 byte = 0x18
const TAGTYPE_STR9 byte = 0x19
const TAGTYPE_STR10 byte = 0x1A
const TAGTYPE_STR11 byte = 0x1B
const TAGTYPE_STR12 byte = 0x1C
const TAGTYPE_STR13 byte = 0x1D
const TAGTYPE_STR14 byte = 0x1E
const TAGTYPE_STR15 byte = 0x1F
const TAGTYPE_STR16 byte = 0x20
const TAGTYPE_STR17 byte = 0x21 // accepted by eMule 0.42f (02-Mai-2004) in receiving code
// only because of a flaw, those tags are handled correctly,
// but should not be handled at all
const TAGTYPE_STR18 byte = 0x22 // accepted by eMule 0.42f (02-Mai-2004) in receiving code
//  only because of a flaw, those tags are handled correctly,
// but should not be handled at all
const TAGTYPE_STR19 byte = 0x23 // accepted by eMule 0.42f (02-Mai-2004) in receiving code
// only because of a flaw, those tags are handled correctly,
// but should not be handled at all
const TAGTYPE_STR20 byte = 0x24 // accepted by eMule 0.42f (02-Mai-2004) in receiving code
// only because of a flaw, those tags are handled correctly,
// but should not be handled at all
const TAGTYPE_STR21 byte = 0x25 // accepted by eMule 0.42f (02-Mai-2004) in receiving code
// only because of a flaw, those tags are handled correctly,
// but should not be handled at all
const TAGTYPE_STR22 byte = 0x26

const FT_UNDEFINED byte = 0x00         // undefined tag
const FT_FILENAME byte = 0x01          // <string>
const FT_FILESIZE byte = 0x02          // <uint32>
const FT_FILESIZE_HI byte = 0x3A       // <uint32>
const FT_FILETYPE byte = 0x03          // <string> or <uint32>
const FT_FILEFORMAT byte = 0x04        // <string>
const FT_LASTSEENCOMPLETE byte = 0x05  // <uint32>
const FT_TRANSFERRED byte = 0x08       // <uint32>
const FT_GAPSTART byte = 0x09          // <uint32>
const FT_GAPEND byte = 0x0A            // <uint32>
const FT_PARTFILENAME byte = 0x12      // <string>
const FT_OLDDLPRIORITY byte = 0x13     // Not used anymore
const FT_STATUS byte = 0x14            // <uint32>
const FT_SOURCES byte = 0x15           // <uint32>
const FT_PERMISSIONS byte = 0x16       // <uint32>
const FT_OLDULPRIORITY byte = 0x17     // Not used anymore
const FT_DLPRIORITY byte = 0x18        // Was 13
const FT_ULPRIORITY byte = 0x19        // Was 17
const FT_KADLASTPUBLISHKEY byte = 0x20 // <uint32>
const FT_KADLASTPUBLISHSRC byte = 0x21 // <uint32>
const FT_FLAGS byte = 0x22             // <uint32>
const FT_DL_ACTIVE_TIME byte = 0x23    // <uint32>
const FT_CORRUPTEDPARTS byte = 0x24    // <string>
const FT_DL_PREVIEW byte = 0x25
const FT_KADLASTPUBLISHNOTES byte = 0x26 // <uint32>
const FT_AICH_HASH byte = 0x27
const FT_FILEHASH byte = 0x28
const FT_COMPLETE_SOURCES byte = 0x30 // nr. of sources which share a
const FT_FAST_RESUME_DATA byte = 0x31 // fast resume data array

const FT_PUBLISHINFO byte = 0x33     // <uint32>
const FT_ATTRANSFERRED byte = 0x50   // <uint32>
const FT_ATREQUESTED byte = 0x51     // <uint32>
const FT_ATACCEPTED byte = 0x52      // <uint32>
const FT_CATEGORY byte = 0x53        // <uint32>
const FT_ATTRANSFERREDHI byte = 0x54 // <uint32>
const FT_MEDIA_ARTIST byte = 0xD0    // <string>
const FT_MEDIA_ALBUM byte = 0xD1     // <string>
const FT_MEDIA_TITLE byte = 0xD2     // <string>
const FT_MEDIA_LENGTH byte = 0xD3    // <uint32> !!!
const FT_MEDIA_BITRATE byte = 0xD4   // <uint32>
const FT_MEDIA_CODEC byte = 0xD5     // <string>
const FT_FILERATING byte = 0xF7      // <uint8>

const CT_NAME byte = 0x01
const CT_SERVER_UDPSEARCH_FLAGS byte = 0x0E
const CT_PORT byte = 0x0F
const CT_VERSION byte = 0x11
const CT_SERVER_FLAGS byte = 0x20 // currently only used to inform a server about supported features
const CT_EMULECOMPAT_OPTIONS byte = 0xEF
const CT_EMULE_RESERVED1 byte = 0xF0
const CT_EMULE_RESERVED2 byte = 0xF1
const CT_EMULE_RESERVED3 byte = 0xF2
const CT_EMULE_RESERVED4 byte = 0xF3
const CT_EMULE_RESERVED5 byte = 0xF4
const CT_EMULE_RESERVED6 byte = 0xF5
const CT_EMULE_RESERVED7 byte = 0xF6
const CT_EMULE_RESERVED8 byte = 0xF7
const CT_EMULE_RESERVED9 byte = 0xF8
const CT_EMULE_UDPPORTS byte = 0xF9
const CT_EMULE_MISCOPTIONS1 byte = 0xFA
const CT_EMULE_VERSION byte = 0xFB
const CT_EMULE_BUDDYIP byte = 0xFC
const CT_EMULE_BUDDYUDP byte = 0xFD
const CT_EMULE_MISCOPTIONS2 byte = 0xFE
const CT_EMULE_RESERVED13 byte = 0xFF
const CT_MOD_VERSION byte = 0x55

const ET_COMPRESSION byte = 0x20
const ET_UDPPORT byte = 0x21
const ET_UDPVER byte = 0x22
const ET_SOURCEEXCHANGE byte = 0x23
const ET_COMMENTS byte = 0x24
const ET_EXTENDEDREQUEST byte = 0x25
const ET_COMPATIBLECLIENT byte = 0x26
const ET_FEATURES byte = 0x27
const ET_MOD_VERSION byte = CT_MOD_VERSION

const ST_SERVERNAME byte = 0x01 // <string>
// Unused (0x02-0x0A)
const ST_DESCRIPTION byte = 0x0B // <string>
const ST_PING byte = 0x0C        // <uint32>
const ST_FAIL byte = 0x0D        // <uint32>
const ST_PREFERENCE byte = 0x0E  // <uint32>
// Unused (0x0F-0x84)
const ST_DYNIP byte = 0x85
const ST_LASTPING_DEPRECATED byte = 0x86 // <uint32> // DEPRECATED, use 0x90
const ST_MAXUSERS byte = 0x87
const ST_SOFTFILES byte = 0x88
const ST_HARDFILES byte = 0x89

// Unused (0x8A-0x8F)
const ST_LASTPING byte = 0x90           // <uint32>
const ST_VERSION byte = 0x91            // <string>
const ST_UDPFLAGS byte = 0x92           // <uint32>
const ST_AUXPORTSLIST byte = 0x93       // <string>
const ST_LOWIDUSERS byte = 0x94         // <uint32>
const ST_UDPKEY byte = 0x95             // <uint32>
const ST_UDPKEYIP byte = 0x96           // <uint32>
const ST_TCPPORTOBFUSCATION byte = 0x97 // <uint16>
const ST_UDPPORTOBFUSCATION byte = 0x98 // <uint16>

// Kad search + some unused tags to mirror the ed2k ones.
const TAG_FILENAME byte = 0x01    // <string>
const TAG_FILESIZE byte = 0x02    // <uint32>
const TAG_FILESIZE_HI byte = 0x3A // <uint32>
const TAG_FILETYPE byte = 0x03    // <string>
const TAG_FILEFORMAT byte = 0x04  // <string>
const TAG_COLLECTION byte = 0x05
const TAG_PART_PATH byte = 0x06 // <string>
const TAG_PART_HASH byte = 0x07
const TAG_COPIED byte = 0x08      // <uint32>
const TAG_GAP_START byte = 0x09   // <uint32>
const TAG_GAP_END byte = 0x0A     // <uint32>
const TAG_DESCRIPTION byte = 0x0B // <string>
const TAG_PING byte = 0x0C
const TAG_FAIL byte = 0x0D
const TAG_PREFERENCE byte = 0x0E
const TAG_PORT byte = 0x0F
const TAG_IP_ADDRESS byte = 0x10
const TAG_VERSION byte = 0x11      // <string>
const TAG_TEMPFILE byte = 0x12     // <string>
const TAG_PRIORITY byte = 0x13     // <uint32>
const TAG_STATUS byte = 0x14       // <uint32>
const TAG_SOURCES byte = 0x15      // <uint32>
const TAG_AVAILABILITY byte = 0x15 // <uint32>
const TAG_PERMISSIONS byte = 0x16
const TAG_QTIME byte = 0x16
const TAG_PARTS byte = 0x17
const TAG_PUBLISHINFO byte = 0x33    // <uint32>
const TAG_MEDIA_ARTIST byte = 0xD0   // <string>
const TAG_MEDIA_ALBUM byte = 0xD1    // <string>
const TAG_MEDIA_TITLE byte = 0xD2    // <string>
const TAG_MEDIA_LENGTH byte = 0xD3   // <uint32> !!!
const TAG_MEDIA_BITRATE byte = 0xD4  // <uint32>
const TAG_MEDIA_CODEC byte = 0xD5    // <string>
const TAG_KADMISCOPTIONS byte = 0xF2 // <uint8>
const TAG_ENCRYPTION byte = 0xF3     // <uint8>
const TAG_FILERATING byte = 0xF7     // <uint8>
const TAG_BUDDYHASH byte = 0xF8      // <string>
const TAG_CLIENTLOWID byte = 0xF9    // <uint32>
const TAG_SERVERPORT byte = 0xFA     // <uint16>
const TAG_SERVERIP byte = 0xFB       // <uint32>
const TAG_SOURCEUPORT byte = 0xFC    // <uint16>
const TAG_SOURCEPORT byte = 0xFD     // <uint16>
const TAG_SOURCEIP byte = 0xFE       // <uint32>
const TAG_SOURCETYPE byte = 0xFF     // <uint8>

func TagType2String(id byte) string {
	switch id {
	case TAGTYPE_UNDEFINED:
		return "TAGTYPE_UNDEFINED"
	case TAGTYPE_HASH16:
		return "TAGTYPE_HASH16"
	case TAGTYPE_STRING:
		return "TAGTYPE_STRING"
	case TAGTYPE_UINT32:
		return "TAGTYPE_UINT32"
	case TAGTYPE_FLOAT32:
		return "TAGTYPE_FLOAT32"
	case TAGTYPE_BOOL:
		return "TAGTYPE_BOOL"
	case TAGTYPE_BOOLARRAY:
		return "TAGTYPE_BOOLARRAY"
	case TAGTYPE_BLOB:
		return "TAGTYPE_BLOB"
	case TAGTYPE_UINT16:
		return "TAGTYPE_UINT16"
	case TAGTYPE_UINT8:
		return "TAGTYPE_UINT8"
	case TAGTYPE_BSOB:
		return "TAGTYPE_BSOB"
	case TAGTYPE_UINT64:
		return "TAGTYPE_UINT64"

	case TAGTYPE_STR1:
		return "TAGTYPE_STR1"
	case TAGTYPE_STR2:
		return "TAGTYPE_STR2"
	case TAGTYPE_STR3:
		return "TAGTYPE_STR3"
	case TAGTYPE_STR4:
		return "TAGTYPE_STR4"
	case TAGTYPE_STR5:
		return "TAGTYPE_STR5"
	case TAGTYPE_STR6:
		return "TAGTYPE_STR6"
	case TAGTYPE_STR7:
		return "TAGTYPE_STR7"
	case TAGTYPE_STR8:
		return "TAGTYPE_STR8"
	case TAGTYPE_STR9:
		return "TAGTYPE_STR9"
	case TAGTYPE_STR10:
		return "TAGTYPE_STR10"
	case TAGTYPE_STR11:
		return "TAGTYPE_STR11"
	case TAGTYPE_STR12:
		return "TAGTYPE_STR12"
	case TAGTYPE_STR13:
		return "TAGTYPE_STR13"
	case TAGTYPE_STR14:
		return "TAGTYPE_STR14"
	case TAGTYPE_STR15:
		return "TAGTYPE_STR15"
	case TAGTYPE_STR16:
		return "TAGTYPE_STR16"
	case TAGTYPE_STR17:
		return "TAGTYPE_STR17"
	case TAGTYPE_STR18:
		return "TAGTYPE_STR18"
	case TAGTYPE_STR19:
		return "TAGTYPE_STR19"
	case TAGTYPE_STR20:
		return "TAGTYPE_STR20"
	case TAGTYPE_STR21:
		return "TAGTYPE_STR21"
	case TAGTYPE_STR22:
		return "TAGTYPE_STR22"
	default:
		return fmt.Sprintf("%x", id)
	}
}

func TagId2String(id byte) string {
	switch id {
	case FT_UNDEFINED:
		return "FT_UNDEFINED"
	case FT_FILENAME:
		return "FT_FILENAME/CT_NAME"
	case FT_FILESIZE:
		return "FT_FILESIZE"
	case FT_FILESIZE_HI:
		return "FT_FILESIZE_HI"
	case FT_FILETYPE:
		return "FT_FILETYPE"
	case FT_FILEFORMAT:
		return "FT_FILEFORMAT"
	case FT_LASTSEENCOMPLETE:
		return "FT_LASTSEENCOMPLETE"
	case FT_TRANSFERRED:
		return "FT_TRANSFERRED"
	case FT_GAPSTART:
		return "FT_GAPSTART"
	case FT_GAPEND:
		return "FT_GAPEND"
	case FT_PARTFILENAME:
		return "FT_PARTFILENAME"
	case FT_OLDDLPRIORITY:
		return "FT_OLDDLPRIORITY"
	case FT_STATUS:
		return "FT_STATUS"
	case FT_SOURCES:
		return "FT_SOURCES"
	case FT_PERMISSIONS:
		return "FT_PERMISSIONS"
	case FT_OLDULPRIORITY:
		return "FT_OLDULPRIORITY"
	case FT_DLPRIORITY:
		return "FT_DLPRIORITY"
	case FT_ULPRIORITY:
		return "FT_ULPRIORITY"
	case FT_KADLASTPUBLISHKEY:
		return "FT_KADLASTPUBLISHKEY/CT_SERVER_FLAGS"
	case FT_KADLASTPUBLISHSRC:
		return "FT_KADLASTPUBLISHSRC"
	case FT_FLAGS:
		return "FT_FLAGS"
	case FT_DL_ACTIVE_TIME:
		return "FT_DL_ACTIVE_TIME"
	case FT_CORRUPTEDPARTS:
		return "FT_CORRUPTEDPARTS"
	case FT_DL_PREVIEW:
		return "FT_DL_PREVIEW"
	case FT_KADLASTPUBLISHNOTES:
		return "FT_KADLASTPUBLISHNOTES"
	case FT_AICH_HASH:
		return "FT_AICH_HASH"
	case FT_FILEHASH:
		return "FT_FILEHASH"
	case FT_COMPLETE_SOURCES:
		return "FT_COMPLETE_SOURCES"
	case FT_FAST_RESUME_DATA:
		return "FT_FAST_RESUME_DATA"
	case CT_SERVER_UDPSEARCH_FLAGS:
		return "CT_SERVER_UDPSEARCH_FLAGS"
	case CT_PORT:
		return "CT_PORT"
	case CT_VERSION:
		return "CT_VERSION"
	case CT_EMULECOMPAT_OPTIONS:
		return "CT_EMULECOMPAT_OPTIONS"
	case CT_EMULE_RESERVED1:
		return "CT_EMULE_RESERVED1"
	case CT_EMULE_RESERVED2:
		return "CT_EMULE_RESERVED2"
	case CT_EMULE_RESERVED3:
		return "CT_EMULE_RESERVED3"
	case CT_EMULE_RESERVED4:
		return "[CT_EMULE_RESERVED4/TAG_ENCRYPTION]"
	case CT_EMULE_RESERVED5:
		return "CT_EMULE_RESERVED5"
	case CT_EMULE_RESERVED6:
		return "CT_EMULE_RESERVED6"
	case CT_EMULE_RESERVED7:
		return "CT_EMULE_RESERVED7"
	case CT_EMULE_RESERVED8:
		return "CT_EMULE_RESERVED8"
	case CT_EMULE_RESERVED9:
		return "[CT_EMULE_RESERVED9/TAG_BUDDYHASH]"
	case CT_EMULE_UDPPORTS:
		return "[CT_EMULE_UDPPORTS/TAG_CLIENTLOWID]"
	case CT_EMULE_MISCOPTIONS1:
		return "[CT_EMULE_MISCOPTIONS1/TAG_SERVERPORT]"
	case CT_EMULE_VERSION:
		return "[CT_EMULE_VERSION/TAG_SERVERIP]"
	case CT_EMULE_BUDDYIP:
		return "[CT_EMULE_BUDDYIP/TAG_SOURCEUPORT]"
	case CT_EMULE_BUDDYUDP:
		return "[CT_EMULE_BUDDYUDP/TAG_SOURCEPORT]"
	case CT_EMULE_MISCOPTIONS2:
		return "[CT_EMULE_MISCOPTIONS2/TAG_SOURCEIP]"
	case CT_EMULE_RESERVED13:
		return "[CT_EMULE_RESERVED13/TAG_SOURCETYPE]"
	case CT_MOD_VERSION:
		return "CT_MOD_VERSION"
	default:
		return fmt.Sprintf("%v", id)
	}
}

type Tag struct {
	Type  byte
	Id    byte
	Name  string
	value []byte
}

func (t *Tag) Get(sr *StateBuffer) *StateBuffer {
	sr.Read(&t.Type)
	if sr.err == nil {
		if (t.Type & 0x80) != 0 {
			t.Type &= 0x7f
			sr.Read(&t.Id)
		} else {
			var l uint16
			sr.Read(&l)
			if sr.err == nil && int(l) < 100500 {
				bc := make([]byte, l)
				sr.Read(bc)
				if l == 1 {
					t.Id = bc[0]
				} else {
					t.Name = string(bc)
				}
			}
		}
	}

	var bc uint32 = 0
	switch {
	case t.Type == TAGTYPE_UINT8:
		bc = 1
	case t.Type == TAGTYPE_UINT16:
		bc = 2
	case t.Type == TAGTYPE_UINT32:
		bc = 4
	case t.Type == TAGTYPE_UINT64:
		bc = 8
	case t.Type == TAGTYPE_FLOAT32:
		bc = 4
	case t.Type == TAGTYPE_BOOL:
		bc = 1
	case t.Type >= TAGTYPE_STR1 && t.Type <= TAGTYPE_STR16:
		bc = uint32(t.Type - TAGTYPE_STR1 + 1)
	case t.Type == TAGTYPE_STRING:
		var v uint16
		sr.Read(&v)
		bc = uint32(v)
	case t.Type == TAGTYPE_BLOB:
		var v uint32
		sr.Read(&v)
		bc = v
	case t.Type == TAGTYPE_BSOB:
		var v byte
		sr.Read(&v)
		bc = uint32(v)
	case t.Type == TAGTYPE_BOOLARRAY:
		var v uint16
		sr.Read(&v)
		bc = uint32(v)
	case t.Type == TAGTYPE_HASH16:
		bc = 16
	default:
		sr.err = errors.New("Tag Get unknown type " + fmt.Sprintf("%x", t.Type))
		return sr
	}

	if sr.err == nil && bc > 0 && bc < 100000 {
		t.value = make([]byte, bc)
		sr.Read(t.value)
	}

	return sr
}

func (t Tag) Put(sw *StateBuffer) *StateBuffer {
	if sw.err != nil {
		return sw
	}

	if t.Name == "" {
		sw.Write((byte)(t.Type | 0x80)).Write(t.Id)
	} else {
		bc := []byte(t.Name)
		sw.Write(t.Type).Write(uint16(len(t.Name))).Write(bc)
	}

	switch {
	case t.Type == TAGTYPE_UINT8 || t.Type == TAGTYPE_UINT16 ||
		t.Type == TAGTYPE_UINT32 || t.Type == TAGTYPE_UINT64 ||
		t.Type == TAGTYPE_FLOAT32 || t.Type == TAGTYPE_BOOL ||
		t.Type == TAGTYPE_HASH16 ||
		(t.Type >= TAGTYPE_STR1 && t.Type <= TAGTYPE_STR16):
		sw.Write(t.value)
	case t.Type == TAGTYPE_STRING || t.Type == TAGTYPE_BOOLARRAY:
		sw.Write(uint16(len(t.value))).Write(t.value)
	case t.Type == TAGTYPE_BSOB:
		sw.Write(byte(len(t.value))).Write(t.value)
	case t.Type == TAGTYPE_BLOB:
		sw.Write(uint32(len(t.value))).Write(t.value)
	default:
		sw.err = errors.New("Tag Put unknown type " + fmt.Sprintf("%x", t.Type))
	}
	return sw
}

func (t Tag) Size() int {
	res := 0
	if t.Name == "" {
		res += DataSize(t.Type) + DataSize(t.Id)
	} else {
		bc := []byte(t.Name)
		res += DataSize(t.Type) + DataSize(uint16(len(t.Name))) + DataSize(bc)
	}

	switch {
	case t.Type == TAGTYPE_UINT8 || t.Type == TAGTYPE_UINT16 ||
		t.Type == TAGTYPE_UINT32 || t.Type == TAGTYPE_UINT64 ||
		t.Type == TAGTYPE_FLOAT32 || t.Type == TAGTYPE_BOOL ||
		t.Type == TAGTYPE_HASH16 ||
		(t.Type >= TAGTYPE_STR1 && t.Type <= TAGTYPE_STR16):
		res += DataSize(t.value)
	case t.Type == TAGTYPE_STRING || t.Type == TAGTYPE_BOOLARRAY:
		res += DataSize(uint16(len(t.value))) + DataSize(t.value)
	case t.Type == TAGTYPE_BSOB:
		res += DataSize(byte(len(t.value))) + DataSize(t.value)
	case t.Type == TAGTYPE_BLOB:
		res += DataSize(uint32(len(t.value))) + DataSize(t.value)
	default:
		res = -1 // raise exception
	}
	return res
}

func (t Tag) GetName() string {
	return t.Name
}

func (t Tag) IsByte() bool {
	return t.Type == TAGTYPE_UINT8
}

func (t Tag) AsByte() byte {
	return t.value[0]
}

func (t Tag) IsUint16() bool {
	return t.Type == TAGTYPE_UINT16
}

func (t Tag) AsUint16() uint16 {
	return binary.LittleEndian.Uint16(t.value)
}

func (t Tag) IsUint32() bool {
	return t.Type == TAGTYPE_UINT32
}

func (t Tag) AsUint32() uint32 {
	return binary.BigEndian.Uint32(t.value)
}

func (t Tag) IsUint64() bool {
	return t.Type == TAGTYPE_UINT64
}

func (t Tag) AsUint64() uint64 {
	return binary.LittleEndian.Uint64(t.value)
}

func (t Tag) IsString() bool {
	return (t.Type >= TAGTYPE_STR1 && t.Type <= TAGTYPE_STR16) || t.Type == TAGTYPE_STRING
}

func (t Tag) AsString() string {
	return string(t.value)
}

func (t Tag) IsBool() bool {
	return t.Type == TAGTYPE_BOOL
}

func (t Tag) AsBool() bool {
	return t.value[0] != 0x00
}

func (t Tag) IsBlob() bool {
	return t.Type == TAGTYPE_BLOB
}

func (t Tag) AsBlob() []byte {
	return t.value
}

func (t Tag) IsHash() bool {
	return t.Type == TAGTYPE_HASH16
}

func (t Tag) AsHash() []byte {
	return t.value
}

func (t Tag) IsFloat() bool {
	return t.Type == TAGTYPE_FLOAT32
}

func (t Tag) AsFloat() float32 {
	x := binary.LittleEndian.Uint32(t.value)
	return math.Float32frombits(x)
}

func (t Tag) AsInt() int {
	switch t.Type {
	case TAGTYPE_UINT8:
		return int(t.AsByte())
	case TAGTYPE_UINT16:
		return int(t.AsUint16())
	case TAGTYPE_UINT32:
		return int(t.AsUint32())
	case TAGTYPE_UINT64:
		return -1
	default:
		return -2
	}
}

func CreateTag(data interface{}, id byte, name string) Tag {
	switch data := data.(type) {
	case byte:
		return Tag{Type: TAGTYPE_UINT8, Id: id, Name: name, value: []byte{byte(data)}}
	case uint16:
		return Tag{Type: TAGTYPE_UINT16, Id: id, Name: name, value: []byte{byte(data & 0xFF), byte(uint16(data) >> 8)}}
	case uint32:
		return Tag{Type: TAGTYPE_UINT32, Id: id, Name: name, value: []byte{byte(data & 0xFF), byte((uint32(data) >> 8) & 0xFF), byte((uint32(data) >> 16) & 0xFF), byte((uint32(data) >> 24) & 0xFF)}}
	case uint64:
		v := make([]byte, 8)
		binary.LittleEndian.PutUint64(v, data)
		return Tag{Type: TAGTYPE_UINT64, Id: id, Name: name, value: v}
	case float32:
		v := math.Float32bits(data)
		return Tag{Type: TAGTYPE_FLOAT32, Id: id, Name: name, value: []byte{byte(v & 0xFF), byte((v >> 8) & 0xFF), byte((v >> 16) & 0xFF), byte((v >> 24) & 0xFF)}}
	case string:
		v := []byte(data)
		if len(v) <= 16 {
			return Tag{Type: TAGTYPE_STR1 + byte(len(v)) - 1, Id: id, Name: name, value: v}
		}

		return Tag{Type: TAGTYPE_STRING, Id: id, Name: name, value: v}
	case bool:
		var b byte
		if data {
			b = 0x01
		}
		return Tag{Type: TAGTYPE_BOOL, Id: id, Name: name, value: []byte{b}}
	case []byte:
		return Tag{Type: TAGTYPE_BLOB, Id: id, Name: name, value: data}
	case Hash:
		return Tag{Type: TAGTYPE_HASH16, Id: id, Name: name, value: data[:]}
	default:
		return Tag{Type: TAGTYPE_UNDEFINED, Id: FT_UNDEFINED, value: []byte{}}
	}
}
