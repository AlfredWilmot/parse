package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
)

var (
	ErrBencoding              = errors.New("invalid bencoding token")
	ErrBencodingNotListOrDict = errors.New("invalid bencoding, must be BencodingList or BencodingDict")
	ErrBencodingByteString    = errors.New("invalid bencodingByteStringType")
	ErrBencodingInt64         = errors.New("invalid bencodingByteInt64")
	ErrBencodingList          = errors.New("invalid bencodingList")
	ErrBencodingDict          = errors.New("invalid bencodingDict")
)

type BencodingType int

const (
	BencodingInvalidType BencodingType = iota
	BencodingByteStringType
	BencodingInt64Type
	BencodingListType
	BencodingDictType
)

// TokenType helps the parser perform the appropriate data-format transformation
// for different data-structures while streaming data.
type TokenType int

const (
	TokenSingleItem TokenType = iota
	TokenKeyValue
	TokenListStart
	TokenListEnd
	TokenDictStart
	TokenDictEnd
)

type Tokener interface {
	TokenData() []byte
	ParsedData() []byte
	String() string
}

// BencodingToken represents a parsed token of bencoded data
type BencodingToken struct {
	parsedData []byte // slice referencing token data portion in buffer (e.g. foo | 3  | i-3ei3e | 3:fooi-3e3:bari3e)
	tokenData  []byte // slice referencing entire token in buffer
	label      BencodingType
}

type (
	BencodingList []BencodingToken
	BencodingDict map[*BencodingToken]BencodingToken
)

func (b BencodingToken) TokenData() []byte {
	return b.tokenData
}

func (b BencodingToken) ParsedData() []byte {
	return b.parsedData
}

// String is a convenience method for returning the tokenData contents as a string.
func (b BencodingToken) String() string {
	return string(b.tokenData)
}

// ParseBencoding returns a collection of BencodingTokens parsed from a buffer.
func ParseBencoding(data []byte) (BencodingToken, error) {
	var err error
	var token BencodingToken

	if len(data) < 2 {
		return token, fmt.Errorf("%v: %v", ErrBencoding, string(data))
	}

	switch {
	case data[0] >= '0' && data[0] <= '9':
		token, err = newBencodingByteString(data)
		if err != nil {
			return token, err
		}
		slog.Info("detected BencodingByteString", "token", token.String())
	case data[0] == 'i':
		token, err = newBencodingInt64(data)
		if err != nil {
			return token, err
		}
		slog.Info("detected BencodingInt64", "token", token.String())
	case data[0] == 'l':
		tail := 0
		token = BencodingToken{nil, nil, BencodingListType}

		for tail < len(data) {

			if data[tail+1] == 'e' {
				token.tokenData = data[:tail+2]
				slog.Info("detected BencodingList", "token", token.String())
				break
			}

			// parse next list element
			subToken, err := ParseBencoding(data[tail+1 : len(data)-1])
			if err != nil {
				return token, err
			}
			tail += len(subToken.TokenData())
		}

	case data[0] == 'd':
		// empty dict
		if data[1] == 'e' {
			token = BencodingToken{nil, data, BencodingDictType}
		}
	default:
		return token, fmt.Errorf("%v: '%v'", ErrBencoding, string(data))
	}
	return token, err
}

// newBencodingByteString attempts to parse data for a valid ByteString Token from the provided buffer.
// Encoding: <string length encoded in base ten ASCII>:<string data>
// (https://wiki.theory.org/BitTorrentSpecification#Byte_Strings)
func newBencodingByteString(data []byte) (BencodingToken, error) {
	// check the start of the ByteString is valid

	if len(data) < 2 {
		return BencodingToken{}, fmt.Errorf("BencodingByteStringType must be at least two characters")
	}

	switch {
	case data[0] == '0' && data[1] == ':':
		// 0-length ByteString represents an empty string
		return BencodingToken{[]byte(""), data[0:2], BencodingByteStringType}, nil
	case data[0] >= '1' && data[0] <= '9':
		// noop (valid starting digits for non-empty strings)
	case data[0] == '0' && data[0] != ':':
		return BencodingToken{}, fmt.Errorf("non-empty ByteString with leading 0 not allowed")
	default:
		return BencodingToken{}, fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[0:1]))
	}

	// ensure all characters up-to ':' delimiter are only digits
	var head uint64
	var tail uint64
	for tail < uint64(len(data)) {
		tail++
		if data[tail] == ':' {
			break
		}
		if data[tail] < '0' || data[tail] > '9' {
			return BencodingToken{}, fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[head:tail+1]))
		}
	}
	if data[tail] != ':' && tail == uint64(len(data))-1 {
		return BencodingToken{}, fmt.Errorf("exhaused input buffer but ByteString delimeter ':' never reached (%s)", string(data[head:tail+1]))
	}

	// determine expected length of string data
	lenData := string(data[head:tail])
	strLen, err := strconv.ParseUint(lenData, 10, 64)
	if err != nil {
		return BencodingToken{}, fmt.Errorf("invalid uint64 (%s): %v", lenData, err)
	}
	// shift head to start of ASCII string data segment
	head = tail + 1

	// return token for Bencoding Byte String if sufficient data for indicated string length
	if availableData := (uint64(len(data)) - (tail + 1)); strLen > availableData {
		return BencodingToken{}, fmt.Errorf(
			"expecting more string data (%d) than is available (%d) in '%s'",
			strLen, availableData, string(data[(tail+1):(tail+availableData+1)]),
		)
	}
	// shift tail to end of ASCII string data segment
	tail += strLen
	// final token references slices for both whole token and data segments
	return BencodingToken{data[head : tail+1], data[:tail+1], BencodingByteStringType}, nil
}

// newBencodingInt64 attempts to parse data for a valid Int64 Token from the provided buffer.
// Encoding: i<integer encoded in base ten ASCII>e
// (https://wiki.theory.org/BitTorrentSpecification#Integers)
func newBencodingInt64(data []byte) (BencodingToken, error) {
	var tail uint64

	// verify first character is int64 start delimiter
	if data[tail] != 'i' {
		return BencodingToken{}, fmt.Errorf("first character of bencodingInt64 must be 'i' (%v)", string(data))
	}

	// verify initial character combos are valid
	tail++
	switch {
	case data[tail] == '-' && (data[tail+1] < '1' || data[tail+1] > '9'):
		return BencodingToken{}, fmt.Errorf("only digits 1-9 can immediately follow a minus symbol (%v)", string(data[:tail+2]))
	case data[tail] == '0' && data[tail+1] == 'e':
		return BencodingToken{data[tail : tail+1], data[:tail+2], BencodingInt64Type}, nil // zero-value bencodingInt64
	case data[tail] == '0' && data[tail+1] != 'e':
		return BencodingToken{}, fmt.Errorf("cannot have a leading '0' that is not immediately terminated with an 'e' (%v)", string(data[:tail+2]))
	case data[tail] != '-' && (data[tail] < '0' || data[tail] > '9'):
		return BencodingToken{}, fmt.Errorf("only digits 0-9 or minus symbol can start a BencodingInt64Type (%v)", string(data[:tail+1]))
	}

	// scan remainder of buffer
	tail++
	for tail < uint64(len(data)) {
		switch {
		case data[tail] == 'e':
			_, err := strconv.ParseInt(string(data[1:tail]), 10, 64)
			if err != nil {
				return BencodingToken{}, fmt.Errorf("could not parse into int64 (%v)", string(data[+1:tail]))
			}
			return BencodingToken{data[1:tail], data[:tail+1], BencodingInt64Type}, nil
		case data[tail] < '0' || data[tail] > '9':
			return BencodingToken{}, fmt.Errorf("illegal character (%c) detected while parsing int64 bytes (%s)", data[tail], string(data[:tail+1]))
		}
		tail++
	}
	return BencodingToken{nil, data[:tail+1], BencodingInvalidType}, fmt.Errorf("exhuasted buffer while parsing as bencodingInt64 (%s)", string(data[:tail+1]))
}
