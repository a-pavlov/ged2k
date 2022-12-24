package proto

import (
	"fmt"
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

func CreateCloseParen() *OperatorEntry {
	x := OperatorEntry(')')
	return &x
}

func CreateOpenParen() *OperatorEntry {
	x := OperatorEntry('(')
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

func (entry StringEntry) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(entry.value).Write(entry.tag)
}

func (entry *StringEntry) Get(sb *StateBuffer) *StateBuffer {
	return sb
}

func (entry OperatorEntry) Put(sb *StateBuffer) *StateBuffer {
	return sb.Write(SEARCH_TYPE_BOOL).Write(byte(entry))
}

func (entry *OperatorEntry) Get(sb *StateBuffer) *StateBuffer {
	return sb
}

func (o OperatorEntry) IsBoolean() bool {
	return byte(o) == OPER_AND || byte(o) == OPER_OR || byte(o) == OPER_NOT
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
		return result, fmt.Errorf("File type too long")
	}

	if len(fileExtension) > SEARCH_REQ_ELEM_LENGTH {
		return result, fmt.Errorf("File ext too long")
	}

	if len(codec) > SEARCH_REQ_ELEM_LENGTH {
		return result, fmt.Errorf("Codec too long")
	}

	if len(value) > SEARCH_REQ_QUERY_LENGTH {
		return result, fmt.Errorf("Search request too long")
	}

	if len(value) == 0 {
		return result, fmt.Errorf("Search request is empty")
	}

	if fileType == ED2KFTSTR_FOLDER {
		// for folders we search emule collections exclude ed2k links - user brackets to correct expr
		result = append(result, CreateOpenParen())
		result = append(result, CreateStringEntry(ED2KFTSTR_EMULECOLLECTION, FT_FILETYPE))
		result = append(result, CreateNot())
		result = append(result, CreateStringEntryNoTag("ED2K:\\"))
		result = append(result, CreateCloseParen())
	} else {
		if len(fileType) != 0 {
			if fileType == ED2KFTSTR_ARCHIVE || fileType == ED2KFTSTR_CDIMAGE {
				result = append(result, CreateStringEntry(ED2KFTSTR_PROGRAM, FT_FILETYPE))
			} else {
				result = append(result, CreateStringEntry(fileType, FT_FILETYPE))
			}
		}

		// if type is not folder - process file parameters now
		if fileType != ED2KFTSTR_EMULECOLLECTION {
			if minSize != 0 {
				result = append(result, CreateNumericEntry(minSize, FT_FILESIZE, ED2K_SEARCH_OP_GREATER))
			}

			if maxSize != 0 {
				result = append(result, CreateNumericEntry(maxSize, FT_FILESIZE, ED2K_SEARCH_OP_LESS))
			}

			if sourcesCount != 0 {
				result = append(result, CreateNumericEntry(uint64(sourcesCount), FT_SOURCES, ED2K_SEARCH_OP_GREATER))
			}

			if completeSourcesCount != 0 {
				result = append(result, CreateNumericEntry(uint64(completeSourcesCount), FT_COMPLETE_SOURCES, ED2K_SEARCH_OP_GREATER))
			}

			if len(fileExtension) != 0 {
				result = append(result, CreateStringEntry(fileExtension, FT_FILEFORMAT))
			}

			if len(codec) != 0 {
				result = append(result, CreateStringEntry(codec, FT_MEDIA_CODEC))
			}

			if mediaLength != 0 {
				result = append(result, CreateNumericEntry(uint64(mediaLength), FT_MEDIA_LENGTH, ED2K_SEARCH_OP_GREATER_EQUAL))
			}

			if mediaBitrate != 0 {
				result = append(result, CreateNumericEntry(uint64(mediaBitrate), FT_MEDIA_BITRATE, ED2K_SEARCH_OP_GREATER_EQUAL))
			}
		}
	}

	verbatim := false
	item := ""

	for _, c := range value {
		switch c {
		case ' ':
		case '(':
		case ')':
			if verbatim {
				item += string(c)
			} else if len(item) != 0 {
				oper := true
				switch item {
				case "AND":
					result = append(result, CreateAnd())
				case "OR":
					result = append(result, CreateOr())
				case "NOT":
					result = append(result, CreateNot())
				default:
					result = append(result, CreateStringEntryNoTag(item))
					oper = false
				}

				if oper {
					if len(result) == 1 {
						return result, fmt.Errorf("Boolean operator at the beginnig of the search expression")
					}

					if _, ok := result[len(result)-1].(*OperatorEntry); ok {
						return result, fmt.Errorf("Two boolean operators in a row")
					}
				}

				item = ""

			}

			if c == '(' {
				result = append(result, CreateOpenParen())
			}

			if c == ')' {
				result = append(result, CreateCloseParen())
			}
		case '"':
			verbatim = !verbatim // change verbatim status and add this character
		default:
			item += string(c)
		}
	}

	return result, nil
}

func PackRequest(source []Serializable) ([]Serializable, error) {
	result := make([]Serializable, 0)
	operators_stack := make([]Serializable, 0)

	for i := len(source) - 1; i >= 0; i-- {
		entry := source[i]

		switch data := entry.(type) {
		case *StringEntry:
			result = append([]Serializable{data}, result...)
		case *NumericEntry:
			result = append([]Serializable{data}, result...)
		case *OperatorEntry:
			if *data == '(' {
				if len(operators_stack) == 0 {
					return result, fmt.Errorf("Incorrect parents count")
				}

				// unroll to first close paren
				for {
					top := len(operators_stack) - 1
					if oper, ok := operators_stack[top].(*OperatorEntry); ok {
						if *oper == ')' {
							break
						}

						result = append([]Serializable{operators_stack[top]}, result...)
						operators_stack = operators_stack[:len(operators_stack)-1]

						if len(operators_stack) == 0 {
							return result, fmt.Errorf("Incorrect parents count 2")
						}

					}
				}

				operators_stack = operators_stack[:len(operators_stack)-1]
			}

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
		}

	}

	if len(operators_stack) != 0 {
		switch data := (operators_stack[0]).(type) {
		case *StringEntry:
			result = append([]Serializable{data}, result...)
		case *NumericEntry:
			result = append([]Serializable{data}, result...)
		case *OperatorEntry:
			if *data == '(' || *data == ')' {
				return result, fmt.Errorf("Incorrect parents count 3")
			}
		}

	}

	return result, nil
}
