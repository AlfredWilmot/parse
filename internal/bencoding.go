package internal

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrBencoding           = errors.New("invalid bencoding")
	ErrBencodingByteString = errors.New("invalid bencodingByteStringType")
	ErrBencodingInt64      = errors.New("invalid bencodingByteInt64")
	ErrBencodingList       = errors.New("invalid bencodingList")
	ErrBencodingDict       = errors.New("invalid bencodingDict")
)

type BencodingType int

const (
	BencodingInvalidType BencodingType = iota
	BencodingByteStringType
	BencodingInt64Type
	BencodingListType
	BencodingDictType
)

type ParsedBencoding interface {
	BencodingToken | ~[]BencodingToken | ~map[*BencodingToken]BencodingToken
}

// BencodingToken represents a parsed token of bencoded data
type BencodingToken struct {
	data  []byte // slice encapsulating parsed token (e.g. foo | 3  | i-3ei3e | 3:fooi-3e3:bari3e)
	label BencodingType
}

// String dumps the entire BencodedToken with delimiters
func (b BencodingToken) String() string {
	switch b.label {
	case BencodingByteStringType:
		return fmt.Sprintf("%d:%v", len(b.data), string(b.data))
	case BencodingInt64Type:
		return fmt.Sprintf("i%ve", string(b.data))
	case BencodingListType:
		return fmt.Sprintf("l%ve", string(b.data))
	case BencodingDictType:
		return fmt.Sprintf("d%ve", string(b.data))
	}
	return ""
}

// ParseBencoding parses the Bencoded data according to the detected token type.
func ParseBencoding(data []byte) (BencodingToken, error) {
	var dataSlice []byte
	var dataType BencodingType // defaults to BencodingInvalidType
	var err error

	if len(data) < 2 {
		return BencodingToken{nil, BencodingInvalidType}, fmt.Errorf("%v: %v", ErrBencoding, string(data))
	}

	prefix := data[0]
	switch {
	case prefix >= '0' && prefix <= '9':
		dataType = BencodingByteStringType
		dataSlice, err = newBencodingByteString(data)
	case prefix == 'i':
		dataType = BencodingInt64Type
		dataSlice, err = newBencodingInt64(data)
	case prefix == 'l':
		// TODO
	case prefix == 'd':
		// TODO
	}
	return BencodingToken{dataSlice, dataType}, err
}

// newBencodingByteString attempts to create a new bencodingByteString instance from the provided buffer.
// Encoding: <string length encoded in base ten ASCII>:<string data>
// (https://wiki.theory.org/BitTorrentSpecification#Byte_Strings)
// examples:
// [0], [:]
// [1-9] [0-9], [:], [ASCII]
// Returns (SliceOfParsedBytes, Error)
func newBencodingByteString(data []byte) ([]byte, error) {
	// check the start of the ByteString is valid
	switch {
	case data[0] == '0' && data[1] == ':':
		return []byte(""), nil // 0-length ByteString represents an empty string
	case data[0] >= '1' && data[0] <= '9':
		// noop (valid starting digits for non-empty strings)
	case data[0] == '0' && data[1] != ':':
		return nil, fmt.Errorf("non-empty ByteString with leading 0 not allowed")
	default:
		return nil, fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[:1]))
	}

	// ensure all characters up-to ':' delimiter are only digits
	var cursor uint64 = 1
readByteStringLengthSegment:
	for {
		char := data[cursor]
		switch {
		case char >= '0' && char <= '9':
			cursor++
		case char == ':':
			break readByteStringLengthSegment
		default:
			return nil, fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[:cursor+1]))
		}
		if data[cursor] != ':' && cursor == uint64(len(data)-1) {
			return nil, fmt.Errorf("exhaused input buffer but ByteString delimeter ':' never reached (%s)", string(data))
		}
	}

	// determine expected length of string data
	lenData := string(data[:cursor])
	strLen, err := strconv.ParseUint(lenData, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid uint64 (%s): %v", lenData, err)
	}

	// return token for Bencoding Byte String if sufficient data for indicated string length
	if availableData := uint64(len(data)) - (cursor + 1); strLen > availableData {
		return nil, fmt.Errorf(
			"expecting more string data (%d) than is available (%d) in '%s'",
			strLen, availableData, string(data[cursor+1:cursor+availableData+1]),
		)
	}
	return data[strLen-1 : cursor+strLen+1], nil
}

// -------------------------------------------------------------------------- //

func newBencodingInt64(data []byte) ([]byte, error) {
	// verify first character is int64 start delimiter
	cursor := 0
	if data[cursor] != 'i' {
		return data[:cursor], fmt.Errorf("first character of bencodingInt64 must be 'i' (%v)", string(data[:cursor]))
	}
	// verify initial character combos are valid
	cursor++
	switch {
	case data[cursor] == '-' && (data[cursor+1] < '1' || data[cursor+1] > '9'):
		return data[:cursor], fmt.Errorf("only digits 1-9 can immediately follow a minus symbol (%v)", string(data[:cursor+2]))
	case data[cursor] == '0' && data[cursor+1] == 'e':
		return data[:cursor], nil // empty bencodingInt64
	case data[cursor] == '0' && data[cursor+1] == '0':
		return data[:cursor], fmt.Errorf("cannot have two leading '0' characters (%v)", string(data[:cursor+2]))
	case data[cursor] != '-' && (data[cursor] < '0' || data[cursor] > '9'):
		return data[:cursor], fmt.Errorf("only digits 0-9 or minus symbol can start a BencodingInt64Type (%v)", string(data[:cursor+1]))
	}

	// scan remainder of buffer
	cursor++
	for cursor < len(data) {
		b := data[cursor]
		switch {
		case b == 'e':
			_, err := strconv.ParseInt(string(data[1:cursor]), 10, 64)
			if err != nil {
				return data[1:cursor], fmt.Errorf("could not parse %s into int64", string(data))
			}
			return data[1:cursor], nil
		case data[cursor] < '0' || data[cursor] > '9':
			return data[:cursor], fmt.Errorf("illegal character (%c) detected while parsing int64 bytes (%s)", b, string(data[:cursor+1]))
		}
		cursor++
	}

	return data[:cursor], fmt.Errorf("exhuasted buffer while parsing as bencodingInt64 (%s)", string(data))
}
