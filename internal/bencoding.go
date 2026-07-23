package internal

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrBencoding           = errors.New("invalid bencoding token")
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

type Tokener interface {
	Token() []byte
	String() string
}

// BencodingToken represents a parsed token of bencoded data
type BencodingToken struct {
	data          []byte // slice encapsulating parsed token (e.g. foo | 3  | i-3ei3e | 3:fooi-3e3:bari3e)
	label         BencodingType
	subTokenCount int // how many tokens are contained within this one (e.g entry in list, or k/v pairs in dict)
}

func (b BencodingToken) Token() []byte {
	return b.data
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

// ParseBencoding returns a collection of BencodingTokens parsed from a buffer.
func ParseBencoding(data []byte, maxTokens int) (chan Tokener, error) {
	var dataSlice []byte
	var err error
	var head, tail uint64

	tokenFeed := make(chan Tokener, maxTokens)

	if len(data) < 2 {
		return nil, fmt.Errorf("%v: %v", ErrBencoding, string(data))
	}

	for head < uint64(len(data)) {
		prefix := data[head]
		switch {
		case prefix >= '0' && prefix <= '9':
			dataSlice, tail, err = newBencodingByteString(head, data)
			if err != nil {
				return nil, err
			}
			tokenFeed <- BencodingToken{dataSlice, BencodingByteStringType, 0}
		case prefix == 'i':
			dataSlice, err = newBencodingInt64(data)
			if err != nil {
				return nil, err
			}
			tokenFeed <- BencodingToken{dataSlice, BencodingInt64Type, 0}
		case prefix == 'l':
			// TODO
		case prefix == 'd':
			// TODO
		default:
			return tokenFeed, fmt.Errorf("%v: '%v'", ErrBencoding, string(data[head:tail+1]))
		}
		tail++
		head = tail
	}
	close(tokenFeed)
	return tokenFeed, err
}

// newBencodingByteString attempts to parse data for a valid ByteString Token from the provided buffer.
// Encoding: <string length encoded in base ten ASCII>:<string data>
// (https://wiki.theory.org/BitTorrentSpecification#Byte_Strings)
func newBencodingByteString(tail uint64, data []byte) ([]byte, uint64, error) {
	// check the start of the ByteString is valid
	head := tail
	switch {
	case data[head] == '0' && data[head+1] == ':':
		return []byte(""), (head + 1), nil // 0-length ByteString represents an empty string
	case data[head] >= '1' && data[head] <= '9':
		// noop (valid starting digits for non-empty strings)
	case data[head] == '0' && data[head] != ':':
		return nil, head, fmt.Errorf("non-empty ByteString with leading 0 not allowed")
	default:
		return nil, head, fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[head:tail+1]))
	}

	// ensure all characters up-to ':' delimiter are only digits
readByteStringLengthSegment:
	for tail < uint64(len(data)) {
		tail++
		switch {
		case data[tail] == ':':
			break readByteStringLengthSegment
		case data[tail] < '0' || data[tail] > '9':
			return nil, tail, fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[head:tail+1]))
		}
	}
	if data[tail] != ':' && tail == uint64(len(data))-1 {
		return nil, tail, fmt.Errorf("exhaused input buffer but ByteString delimeter ':' never reached (%s)", string(data[head:tail+1]))
	}

	// determine expected length of string data
	lenData := string(data[head:tail])
	strLen, err := strconv.ParseUint(lenData, 10, 64)
	if err != nil {
		return nil, tail, fmt.Errorf("invalid uint64 (%s): %v", lenData, err)
	}
	// shift head to start of ASCII string data segment
	head = tail + 1

	// return token for Bencoding Byte String if sufficient data for indicated string length
	if availableData := (uint64(len(data)) - (tail + 1)); strLen > availableData {
		return nil, tail, fmt.Errorf(
			"expecting more string data (%d) than is available (%d) in '%s'",
			strLen, availableData, string(data[(tail+1):(tail+availableData+1)]),
		)
	}
	// shift tail to end of ASCII string data segment
	tail += strLen
	// return segment containing ASCII string data, and final tail position
	return data[head : tail+1], tail, nil
}

// newBencodingInt64 attempts to parse data for a valid Int64 Token from the provided buffer.
// Encoding: i<integer encoded in base ten ASCII>e
// (https://wiki.theory.org/BitTorrentSpecification#Integers)
func newBencodingInt64(data []byte) ([]byte, error) {
	cursor := 0

	// verify first character is int64 start delimiter
	if data[cursor] != 'i' {
		return nil, fmt.Errorf("first character of bencodingInt64 must be 'i' (%v)", string(data[:cursor]))
	}

	// verify initial character combos are valid
	cursor++
	switch {
	case data[cursor] == '-' && (data[cursor+1] < '1' || data[cursor+1] > '9'):
		return nil, fmt.Errorf("only digits 1-9 can immediately follow a minus symbol (%v)", string(data[:cursor+2]))
	case data[cursor] == '0' && data[cursor+1] == 'e':
		return []byte("0"), nil // zero-value bencodingInt64
	case data[cursor] == '0' && data[cursor+1] != 'e':
		return nil, fmt.Errorf("cannot have a leading '0' that is not immediately terminated with an 'e' (%v)", string(data[:cursor+2]))
	case data[cursor] != '-' && (data[cursor] < '0' || data[cursor] > '9'):
		return nil, fmt.Errorf("only digits 0-9 or minus symbol can start a BencodingInt64Type (%v)", string(data[:cursor+1]))
	}

	// scan remainder of buffer
	cursor++
	for cursor < len(data) {
		b := data[cursor]
		switch {
		case b == 'e':
			_, err := strconv.ParseInt(string(data[1:cursor]), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("could not parse into int64 (%v)", string(data))
			}
			return data[1:cursor], nil
		case data[cursor] < '0' || data[cursor] > '9':
			return nil, fmt.Errorf("illegal character (%c) detected while parsing int64 bytes (%s)", b, string(data[:cursor+1]))
		}
		cursor++
	}
	return data[:cursor], fmt.Errorf("exhuasted buffer while parsing as bencodingInt64 (%s)", string(data))
}
