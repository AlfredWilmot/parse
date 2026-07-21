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
	bencodingInt64Type
	bencodingListType
	bencodingDictType
)

type Bencoder interface {
	Type() bencodingType
	String() string
}

type (
	bencodingByteString string
	bencodingInt64      int64
	bencodingList       []Bencoder
	bencodingDict       map[bencodingByteString]Bencoder
)

// -------------------------------------------------------------------------- //

func (b bencodingByteString) String() string {
	return string(b)
}

func (b bencodingByteString) Type() bencodingType {
	return bencodingByteStringType
}

// newBencodingByteString attempts to create a new bencodingByteString instance from the provided buffer.
// Encoding: <string length encoded in base ten ASCII>:<string data>
// (https://wiki.theory.org/BitTorrentSpecification#Byte_Strings)
// examples:
// [0], [:]
// [1-9] [0-9], [:], [ASCII]
func newBencodingByteString(data []byte) (bencodingByteString, error) {
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
			return bencodingByteString(""), fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[:cursor+1]))
		}
		if data[cursor] != ':' && cursor == uint64(len(data)-1) {
			return bencodingByteString(""), fmt.Errorf("exhaused input buffer but ByteString delimeter ':' never reached (%s)", string(data))
		}
	}

	// determine expected length of string data
	lenData := string(data[:cursor])
	strLen, err := strconv.ParseUint(lenData, 10, 64)
	if err != nil {
		return bencodingByteString(""), fmt.Errorf("could not parse '%s', invalid uint64 (%s): %v", lenData, string(data[:cursor]), err)
	}

	// return token for Bencoding Byte String if sufficient data for indicated string length
	if availableData := uint64(len(data)) - (cursor + 1); strLen > availableData {
		err := fmt.Errorf(
			"expecting more string data (%d) than is available (%d) in '%s'",
			strLen, availableData, string(data[cursor+1:cursor+availableData+1]),
		)
		return bencodingByteString(""), err
	}
	return bencodingByteString(string(data[:cursor+strLen+1])), nil
}

// -------------------------------------------------------------------------- //

func (b bencodingInt64) String() string {
	return fmt.Sprintf("i%de", b)
}

func (b bencodingInt64) Type() bencodingType {
	return bencodingInt64Type
}

func newBencodingInt64(data []byte) (bencodingInt64, error) {
	// verify first character is int64 start delimiter
	cursor := 0
	if data[cursor] != 'i' {
		return bencodingInt64(0), fmt.Errorf("first character of bencodingInt64 must be 'i' (%v)", string(data[:cursor]))
	}
	// verify initial character combos are valid
	cursor++
	switch {
	case data[cursor] == '-' && (data[cursor+1] < '1' || data[cursor+1] > '9'):
		return bencodingInt64(0), fmt.Errorf("only digits 1-9 can immediately follow a minus symbol (%v)", string(data[:cursor+2]))
	case data[cursor] == '0' && data[cursor+1] == 'e':
		return bencodingInt64(0), nil // empty bencodingInt64
	case data[cursor] != '-' && (data[cursor] < '0' || data[cursor] > '9'):
		return bencodingInt64(0), fmt.Errorf("only digits 0-9 or minus symbol can start a BencodingInt64Type (%v)", string(data[:cursor+1]))
	}

	// scan remainder of buffer
	cursor++
	for cursor < len(data) {
		b := data[cursor]
		switch {
		case b == 'e':
			vInt64, err := strconv.ParseInt(string(data[1:cursor]), 10, 64)
			if err != nil {
				return bencodingInt64(0), fmt.Errorf("could not parse %s into int64", string(data))
			}
			return bencodingInt64(vInt64), nil
		case data[cursor] < '0' || data[cursor] > '9':
			return bencodingInt64(0), fmt.Errorf("illegal character (%c) detected while parsing int64 bytes (%s)", b, string(data[:cursor+1]))
		}
		cursor++
	}

	return bencodingInt64(0), fmt.Errorf("exhuasted buffer while parsing as bencodingInt64 (%s)", string(data))
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
		return newBencodingByteString(data)
	case prefix == 'i':
		return newBencodingInt64(data)
	case prefix == 'l':
		return nil, nil
	case prefix == 'd':
		return nil, nil
	default:
		return nil, fmt.Errorf("'%c' does not map to the start of a valid bencodingType", prefix)
	}
}
