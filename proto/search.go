package proto

import (
	"fmt"
	"log"
)

const SEARCH_TYPE_BOOL byte = 0x00
const SEARCH_TYPE_STR byte = 0x01
const SEARCH_TYPE_STR_TAG byte = 0x02
const SEARCH_TYPE_UINT32 byte = 0x03
const SEARCH_TYPE_UINT64 byte = 0x08

const SEARCH_REQ_ELEM_LENGTH int = 20
const SEARCH_REQ_QUERY_LENGTH int = 450
const SEARCH_REQ_ELEM_COUNT int = 30

// Media values for FT_FILETYPE
const ED2KFTSTR_AUDIO string = "Audio"
const ED2KFTSTR_VIDEO string = "Video"
const ED2KFTSTR_IMAGE string = "Image"
const ED2KFTSTR_DOCUMENT string = "Doc"
const ED2KFTSTR_PROGRAM string = "Pro"
const ED2KFTSTR_ARCHIVE string = "Arc" // *Mule internal use only
const ED2KFTSTR_CDIMAGE string = "Iso" // *Mule internal use only
const ED2KFTSTR_EMULECOLLECTION string = "EmuleCollection"
const ED2KFTSTR_FOLDER string = "Folder" // Value for eD2K tag FT_FILETYPE
const ED2KFTSTR_USER string = "User"     // eMule internal use only

// Additional media meta data tags from eDonkeyHybrid (note also the uppercase/lowercase)
const FT_ED2K_MEDIA_ARTIST string = "Artist"   // <string>
const FT_ED2K_MEDIA_ALBUM string = "Album"     // <string>
const FT_ED2K_MEDIA_TITLE string = "Title"     // <string>
const FT_ED2K_MEDIA_LENGTH string = "length"   // <string> !!!
const FT_ED2K_MEDIA_BITRATE string = "bitrate" // <uint32>
const FT_ED2K_MEDIA_CODEC string = "codec"     // <string>

const ED2KFT_ANY byte = 0
const ED2KFT_AUDIO byte = 1    // ED2K protocol value (eserver 17.6+)
const ED2KFT_VIDEO byte = 2    // ED2K protocol value (eserver 17.6+)
const ED2KFT_IMAGE byte = 3    // ED2K protocol value (eserver 17.6+)
const ED2KFT_PROGRAM byte = 4  // ED2K protocol value (eserver 17.6+)
const ED2KFT_DOCUMENT byte = 5 // ED2K protocol value (eserver 17.6+)
const ED2KFT_ARCHIVE byte = 6  // ED2K protocol value (eserver 17.6+)
const ED2KFT_CDIMAGE byte = 7  // ED2K protocol value (eserver 17.6+)
const ED2KFT_EMULECOLLECTION byte = 8

const ED2K_SEARCH_OP_EQUAL byte = 0
const ED2K_SEARCH_OP_GREATER byte = 1
const ED2K_SEARCH_OP_LESS byte = 2
const ED2K_SEARCH_OP_GREATER_EQUAL byte = 3
const ED2K_SEARCH_OP_LESS_EQUAL byte = 4
const ED2K_SEARCH_OP_NOTEQUAL byte = 5

const OPER_AND byte = 0x00
const OPER_OR byte = 0x01
const OPER_NOT byte = 0x02

const MaxUint32 = ^uint32(0)

type NumericEntry struct {
	value    uint64
	operator byte
	tag      ByteContainer
}

type StringEntry struct {
	value ByteContainer
	tag   ByteContainer
}

type OperatorEntry byte
type ParenEntry byte

func CreateNumericEntry(val uint64, id byte, op byte) *NumericEntry {
	return &NumericEntry{value: val, operator: op, tag: ByteContainer([]byte{id})}
}

func CreateStringEntry(val string, id byte) *StringEntry {
	return &StringEntry{value: ByteContainer([]byte(val)), tag: ByteContainer([]byte{id})}
}

func CreateStringEntryNoTag(val string) *StringEntry {
	return &StringEntry{value: ByteContainer([]byte(val)), tag: nil}
}

func CreateAnd() *OperatorEntry {
	x := OperatorEntry(OPER_AND)
	return &x
}

func CreateOr() *OperatorEntry {
	x := OperatorEntry(OPER_OR)
	return &x
}

func CreateNot() *OperatorEntry {
	x := OperatorEntry(OPER_NOT)
	return &x
}

func CreateCloseParen() *ParenEntry {
	x := ParenEntry(')')
	return &x
}

func CreateOpenParen() *ParenEntry {
	x := ParenEntry('(')
	return &x
}

func (entry NumericEntry) Put(sb *StateBuffer) *StateBuffer {
	if entry.value < uint64(MaxUint32) {
		sb.Write(uint32(entry.value))
	} else {
		sb.Write(entry.value)
	}
	return sb.Write(entry.operator).Write(entry.tag)
}

func (entry *NumericEntry) Get(sb *StateBuffer) *StateBuffer {
	return sb
}

func (entry NumericEntry) Size() int {
	res := 0
	if entry.value < uint64(MaxUint32) {
		res += DataSize(uint32(entry.value))
	} else {
		res += DataSize(entry.value)
	}

	return res + DataSize(entry.operator) + DataSize(entry.tag)
}

func (entry StringEntry) Put(sb *StateBuffer) *StateBuffer {
	if entry.tag != nil {
		sb.Write(SEARCH_TYPE_STR_TAG)
	} else {
		sb.Write(SEARCH_TYPE_STR)
	}

	sb.Write(entry.value)
	if entry.tag != nil {
		sb.Write(entry.tag)
	}

	return sb
}

func (entry *StringEntry) Get(sb *StateBuffer) *StateBuffer {
	return sb
}

func (entry StringEntry) Size() int {
	if entry.tag != nil {
		return DataSize(SEARCH_TYPE_STR_TAG) + DataSize(entry.value) + DataSize(entry.tag)
	}
	return DataSize(SEARCH_TYPE_STR) + DataSize(entry.value)
}

func (entry OperatorEntry) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(SEARCH_TYPE_BOOL).Write(byte(0))
}

func (entry *OperatorEntry) Get(sb *StateBuffer) *StateBuffer {
	return sb
}

func (entry OperatorEntry) Size() int {
	return DataSize(SEARCH_TYPE_BOOL) + DataSize(byte(entry))
}

func (entry ParenEntry) Put(*StateBuffer) *StateBuffer {
	panic("Requested put for paren entry")
}

func (entry *ParenEntry) Get(*StateBuffer) *StateBuffer {
	panic("Requested get for parent entry")
}

func (entry ParenEntry) Size() uint {
	panic("Requested size for ParenEntry")
}

func (o OperatorEntry) IsBoolean() bool {
	return byte(o) == OPER_AND || byte(o) == OPER_OR || byte(o) == OPER_NOT
}

func (o ParenEntry) IsBoolean() bool {
	return false
}

func addOperand(dst []Serializable, op Serializable) []Serializable {
	_, isOperator := op.(*OperatorEntry)
	srcP, isParen := op.(*ParenEntry)
	isCloseParen := false
	isOpenParen := false
	if isParen {
		if *srcP == '(' {
			isOpenParen = true
		} else {
			isCloseParen = true
		}
	}

	if !isOperator {
		if len(dst) > 0 {
			_, hasOperator := dst[len(dst)-1].(*OperatorEntry)
			dstP, hasParen := dst[len(dst)-1].(*ParenEntry)
			hasCloseParen := false
			hasOpenParen := false
			if hasParen {
				if *dstP == '(' {
					hasOpenParen = true
				} else {
					hasCloseParen = true
				}
			}

			if (!hasParen && !hasOperator && !isParen && !isOperator) || // xxx xxx
				(!hasParen && !hasOperator && isOpenParen) || // xxx (
				(hasCloseParen && !isParen && !isOperator) || // ) xxx
				(hasCloseParen && isOpenParen) { // ) (
				dst = append(dst, CreateAnd())
			}

			if hasOpenParen && isCloseParen {
				// need to report error here
				log.Println("Open-close paren found on addOperand")
			}
		}
	}

	dst = append(dst, op)
	return dst
}

func BuildEntries(minSize uint64,
	maxSize uint64,
	sourcesCount uint32,
	completeSourcesCount uint32,
	fileType string,
	fileExtension string,
	codec string,
	mediaLength uint32,
	mediaBitrate uint32,
	value string) ([]Serializable, error) {
	result := make([]Serializable, 0)

	if len(fileType) > SEARCH_REQ_ELEM_LENGTH {
		return result, fmt.Errorf("file type too long %d", len(fileType))
	}

	if len(fileExtension) > SEARCH_REQ_ELEM_LENGTH {
		return result, fmt.Errorf("file ext too long %d", len(fileExtension))
	}

	if len(codec) > SEARCH_REQ_ELEM_LENGTH {
		return result, fmt.Errorf("codec too long %d", len(codec))
	}

	if len(value) > SEARCH_REQ_QUERY_LENGTH {
		return result, fmt.Errorf("search request too long %d", len(value))
	}

	if len(value) == 0 {
		return result, fmt.Errorf("search request is empty")
	}

	if fileType == ED2KFTSTR_FOLDER {
		// for folders we search emule collections exclude ed2k links - user brackets to correct expr
		result = addOperand(result, CreateOpenParen())
		result = addOperand(result, CreateStringEntry(ED2KFTSTR_EMULECOLLECTION, FT_FILETYPE))
		result = addOperand(result, CreateNot())
		result = addOperand(result, CreateStringEntryNoTag("ED2K:\\"))
		result = addOperand(result, CreateCloseParen())
	} else {
		if len(fileType) != 0 {
			if fileType == ED2KFTSTR_ARCHIVE || fileType == ED2KFTSTR_CDIMAGE {
				result = addOperand(result, CreateStringEntry(ED2KFTSTR_PROGRAM, FT_FILETYPE))
			} else {
				result = addOperand(result, CreateStringEntry(fileType, FT_FILETYPE))
			}
		}

		// if type is not folder - process file parameters now
		if fileType != ED2KFTSTR_EMULECOLLECTION {
			if minSize != 0 {
				result = addOperand(result, CreateNumericEntry(minSize, FT_FILESIZE, ED2K_SEARCH_OP_GREATER))
			}

			if maxSize != 0 {
				result = addOperand(result, CreateNumericEntry(maxSize, FT_FILESIZE, ED2K_SEARCH_OP_LESS))
			}

			if sourcesCount != 0 {
				result = addOperand(result, CreateNumericEntry(uint64(sourcesCount), FT_SOURCES, ED2K_SEARCH_OP_GREATER))
			}

			if completeSourcesCount != 0 {
				result = addOperand(result, CreateNumericEntry(uint64(completeSourcesCount), FT_COMPLETE_SOURCES, ED2K_SEARCH_OP_GREATER))
			}

			if len(fileExtension) != 0 {
				result = addOperand(result, CreateStringEntry(fileExtension, FT_FILEFORMAT))
			}

			if len(codec) != 0 {
				result = addOperand(result, CreateStringEntry(codec, FT_MEDIA_CODEC))
			}

			if mediaLength != 0 {
				result = addOperand(result, CreateNumericEntry(uint64(mediaLength), FT_MEDIA_LENGTH, ED2K_SEARCH_OP_GREATER_EQUAL))
			}

			if mediaBitrate != 0 {
				result = addOperand(result, CreateNumericEntry(uint64(mediaBitrate), FT_MEDIA_BITRATE, ED2K_SEARCH_OP_GREATER_EQUAL))
			}
		}
	}

	verbatim := false
	item := ""

	for _, c := range value {
		switch {
		case c == ' ' || c == '(' || c == ')':
			if verbatim {
				item += string(c)
			} else if len(item) != 0 {
				oper := true
				switch item {
				case "AND":
					result = addOperand(result, CreateAnd())
				case "OR":
					result = addOperand(result, CreateOr())
				case "NOT":
					result = addOperand(result, CreateNot())
				default:
					result = addOperand(result, CreateStringEntryNoTag(item))
					oper = false
				}

				if oper {
					if len(result) == 1 {
						return result, fmt.Errorf("boolean operator at the beginnig of the search expression")
					}

					if _, ok := result[len(result)-2].(*OperatorEntry); ok {
						return result, fmt.Errorf("two boolean operators in a row")
					}
				}

				item = ""

			}

			if c == '(' {
				result = addOperand(result, CreateOpenParen())
			}

			if c == ')' {
				result = addOperand(result, CreateCloseParen())
			}
		case c == '"':
			verbatim = !verbatim // change verbatim status and add this character
		default:
			item += string(c)
		}
	}

	// check unclosed quotes
	if verbatim {
		return result, fmt.Errorf("unclosed quotation mark")
	}

	if len(item) != 0 {
		// add last item - check it is not operator
		if item == "AND" || item == "OR" || item == "NOT" {
			return result, fmt.Errorf("operator at the end of expression")
		}

		result = addOperand(result, CreateStringEntryNoTag(item))
	}

	return result, nil
}

func PackRequest(source []Serializable) (SearchRequest, error) {
	result := SearchRequest{}
	operators_stack := make([]Serializable, 0)

	for i := len(source) - 1; i >= 0; i-- {
		entry := source[i]

		switch data := entry.(type) {
		case *StringEntry:
			result = append([]Serializable{data}, result...)
		case *NumericEntry:
			result = append([]Serializable{data}, result...)
		case *OperatorEntry:
			// we have normal operator and on stack top we have normal operator
			// prepare result - move operator from top to result and replace top
			if data.IsBoolean() && len(operators_stack) > 0 {
				if oper, ok := operators_stack[len(operators_stack)-1].(*OperatorEntry); ok {
					if oper.IsBoolean() {
						result = append([]Serializable{operators_stack[len(operators_stack)-1]}, result...)
						operators_stack = operators_stack[:len(operators_stack)-1]
					}
				}
			}

			operators_stack = append(operators_stack, data)
		case *ParenEntry:
			if *data == '(' {
				if len(operators_stack) == 0 {
					return result, fmt.Errorf("incorrect parents count")
				}

				// unroll to first close paren
			A:
				for {
					top := operators_stack[len(operators_stack)-1]
					oper, ok := top.(*ParenEntry)
					if ok && *oper == ')' {
						break A
					}

					result = append([]Serializable{top}, result...)
					operators_stack = operators_stack[:len(operators_stack)-1]

					if len(operators_stack) == 0 {
						return result, fmt.Errorf("incorrect parents count 2")
					}
				}

				operators_stack = operators_stack[:len(operators_stack)-1]
			} else {
				operators_stack = append(operators_stack, data)
			}
		}

	}

	if len(operators_stack) != 0 {
		switch data := (operators_stack[0]).(type) {
		case *StringEntry:
			result = append([]Serializable{data}, result...)
		case *NumericEntry:
			result = append([]Serializable{data}, result...)
		case *OperatorEntry:
			result = append([]Serializable{data}, result...)
		case *ParenEntry:
			if *data == '(' || *data == ')' {
				return result, fmt.Errorf("incorrect parents count 3")
			}
		}

	}

	return result, nil
}

type SearchRequest []Serializable

func (sr SearchRequest) Put(sb *StateBuffer) *StateBuffer {
	for _, s := range sr {
		s.Put(sb)
	}

	return sb
}

func (sr *SearchRequest) Get(*StateBuffer) *StateBuffer {
	panic("SearchRequest Get issued")
}

func (sr SearchRequest) Size() int {
	res := 0
	for _, s := range sr {
		res += DataSize(s)
	}

	return res
}

type SearchMore struct{}

func (sm SearchMore) Put(sb *StateBuffer) *StateBuffer {
	return sb
}

func (sm *SearchMore) Get(*StateBuffer) *StateBuffer {
	panic("SearchMore Get issued")
}

func (sm SearchMore) Size() int {
	return 0
}

type GetFileSources struct {
	Hash    ED2KHash
	LowPart uint32
	HiPart  uint32
}

func (gfs GetFileSources) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(gfs.Hash).Write(gfs.LowPart).Write(gfs.HiPart)
}

func (gfs *GetFileSources) Get(*StateBuffer) *StateBuffer {
	panic("GetFileSources Get issued")
}

func (gfs GetFileSources) Size() int {
	return DataSize(gfs.Hash) + DataSize(gfs.LowPart) + DataSize(gfs.HiPart)
}

type SearchResult struct {
	Items       []UsualPacket
	MoreResults byte
}

func (sr *SearchResult) Get(sb *StateBuffer) *StateBuffer {
	count := sb.ReadUint32()
	if sb.Error() == nil {
		if count < 1024 {
			sr.Items = make([]UsualPacket, count)
			for i := 0; i < int(count); i++ {
				sr.Items[i].Get(sb)
				if sb.err != nil {
					break
				}
			}

			if sb.Error() == nil && sb.Remain() > 0 {
				sr.MoreResults = sb.ReadUint8()
			}
		} else {
			sb.err = fmt.Errorf("elements count too large")
		}
	}

	return sb
}

func (sr SearchResult) Put(*StateBuffer) *StateBuffer {
	panic("Search result put requested")
}

func (sr SearchResult) Size() int {
	res := DataSize(sr.MoreResults)
	for _, up := range sr.Items {
		res += DataSize(up)
	}

	return res
}

type SearchItem struct {
	H               ED2KHash
	Point           Endpoint
	Filename        string
	Filesize        uint64
	Sources         int
	CompleteSources int
	Bitrate         int
	MediaLength     int
	Codec           string
}

func ToSearchItem(up *UsualPacket) SearchItem {
	res := SearchItem{H: up.Hash, Point: up.Point}
	for _, x := range up.Properties {
		switch x.Id {
		case FT_FILENAME:
			res.Filename = x.AsString()
		case FT_FILESIZE:
			if x.IsUint32() {
				res.Filesize = uint64(x.AsUint32())
			} else if x.IsUint64() {
				res.Filesize = x.AsUint64()
			}
		case FT_SOURCES:
			res.Sources = x.AsInt()
		case FT_COMPLETE_SOURCES:
			res.CompleteSources = x.AsInt()
		case FT_MEDIA_BITRATE:
			res.Bitrate = x.AsInt()
		case FT_MEDIA_LENGTH:
			res.MediaLength = x.AsInt()
		case FT_MEDIA_CODEC:
			res.Codec = x.AsString()
		}
	}
	return res
}
