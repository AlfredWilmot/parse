package internal

import (
	"fmt"
	"strconv"
	"strings"
)

// -------------------------------------------------------------------------- //

type bencodingType int

const (
	bencodingInvalidType bencodingType = iota
	bencodingByteStringType
	bencodingIntegerType
	bencodingListType
	bencodingDictionaryType
)

type Bencoder interface {
	Type() bencodingType
	String() string
}

type (
	bencodingByteString string
	bencodingInteger    int64
	bencodingList       []Bencoder
	bencodingDictionary map[bencodingByteString]Bencoder
)

// -------------------------------------------------------------------------- //

func (b bencodingByteString) String() string {
	return string(b)
}

func (b bencodingByteString) Type() bencodingType {
	return bencodingByteStringType
}

// NewBencodingByteString attempts to create a new bencodingByteString instance from the provided buffer.
// Encoding: <string length encoded in base ten ASCII>:<string data>
// (https://wiki.theory.org/BitTorrentSpecification#Byte_Strings)
// examples:
// [0], [:]
// [1-9] [0-9], [:], [ASCII]
func NewBencodingByteString(data []byte) (bencodingByteString, error) {
	// check the start of the ByteString is valid
	switch {
	case data[0] == '0' && data[1] == ':':
		return bencodingByteString(""), nil // empty ByteString is allowed
	case data[0] >= '1' && data[0] <= '9':
		// noop (valid starting digits for non-empty strings)
	case data[0] == '0' && data[1] != ':':
		return bencodingByteString(""), fmt.Errorf("non-empty ByteString with leading 0 not allowed")
	default:
		return bencodingByteString(""), fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[:1]))
	}

	// ensure all characters up-to ':' delimiter are only digits
	var cursor uint64 = 1
readByteStringLengthSegment:
	for {
		b := data[cursor]
		switch {
		case b >= '0' && b <= '9':
			cursor++
		case b == ':':
			break readByteStringLengthSegment
		default:
			return bencodingByteString(""), fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[:cursor]))
		}
		if data[cursor] != ':' && cursor == uint64(len(data)-1) {
			return bencodingByteString(""), fmt.Errorf("exhaused input buffer but ByteString delimeter ':' never reached (%s)", string(data))
		}
	}

	// determine expected length of string data
	lenData := string(data[:cursor])
	strLen, err := strconv.ParseUint(lenData, 10, 64)
	if err != nil {
		return bencodingByteString(""), fmt.Errorf("could not parse '%s', invalid uint64 (%s)", lenData, string(data[:cursor]))
	}

	// return token for Bencoding Byte String if sufficient data for indicated string length
	availableData := uint64(len(data)) - (cursor + 1)
	if strLen > availableData {
		return bencodingByteString(""), fmt.Errorf("expecting more string data (%d) than is available (%d) in '%s'", strLen, availableData, string(data[cursor+1:cursor+availableData+1]))
	}
	return bencodingByteString(string(data[:cursor+strLen+1])), nil
}

// -------------------------------------------------------------------------- //

func (b bencodingInteger) String() string {
	return fmt.Sprintf("i%de", b)
}

func (b bencodingInteger) Type() bencodingType {
	return bencodingIntegerType
}

// -------------------------------------------------------------------------- //

func (b bencodingList) String() string {
	// initialise this list string
	buff := strings.Builder{}
	buff.WriteString("l")

	// decode each bencoded entry from this list in separate goroutines
	// NOTE: ordering is not guaranteed due to multiple channel senders
	c := make(chan string)
	for _, entry := range b {
		go func() { c <- entry.String() }()
	}
	for str := range c {
		buff.WriteString(str)
	}

	// finalise this list string
	buff.WriteString("e")
	return buff.String()
}

func (b bencodingList) Type() bencodingType {
	return bencodingListType
}

func ParseIntoBencoding(data []byte) (Bencoder, error) {
	prefix := data[0]
	switch {
	case prefix >= '0' && prefix <= '9':
		return NewBencodingByteString(data)
	case prefix == 'i':
		return nil, nil
	case prefix == 'l':
		return nil, nil
	case prefix == 'd':
		return nil, nil
	default:
		return nil, fmt.Errorf("'%c' does not map to the start of a valid bencodingType", prefix)
	}
}

func SplitBencoding(data []byte, atEOF bool) (int, []byte, error) {
	// is the first byte one of the valid starting ASCII characters?
	if b := data[0]; (b < '0' || b > '9') && b != 'i' && b != 'l' && b != 'd' {
		return 0, nil, fmt.Errorf("invalid Bencoding starting character: %c", b)
	}

	// pull data until delimiter for bencodingType reached or buffer is exhausted
	cursor := 1
	for cursor < len(data) {
		switch b := data[cursor]; b {

		// return token for Bencoding values: Int64, List, Dict
		case 'e':
			return cursor, data[:cursor], nil

			// start parsing token as Bencoding Byte String
		case ':':

			// determine expected length of string data
			lenData := string(data[:(cursor - 1)])
			strLen, err := strconv.ParseUint(lenData, 10, 64)
			if err != nil {
				return 0, nil, fmt.Errorf("could not parse '%s', invalid uint64", lenData)
			}

			// return token for Bencoding Byte String if sufficient data for indicated string length
			availableData := uint64(len(data) - cursor)
			if strLen > uint64(len(data)-cursor) {
				return 0, nil, fmt.Errorf("expecting more string data (%d) than is available (%d)", strLen, availableData)
			}
			return cursor, data[:(uint64(cursor) + strLen)], nil
		default:
			cursor++
		}
	}

	// need more data
	return cursor, nil, nil
}
