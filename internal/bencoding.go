package internal

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrBencoding             = errors.New("invalid BencodingToken")
	ErrBencodingByteString   = errors.New("invalid BencodingASCIIString")
	ErrBencodingInt64        = errors.New("invalid BencodingInt64")
	ErrBencodingList         = errors.New("invalid BencodingList")
	ErrBencodingDict         = errors.New("invalid BencodingDict")
	ErrBencodingMissingToken = errors.New("missing BencodingToken")
	ErrBencodingKeyNoValue   = errors.New("key missing corresponding value")
	ErrBencodingKeyInvalid   = errors.New("key must be a BencodingASCIIString")
)

type BencodingType int

const (
	BencodingNull BencodingType = iota
	BencodingASCIIString
	BencodingInt64
	BencodingListStart
	BencodingListEnd
	BencodingDictStart
	BencodingDictEnd
	BencodingDictKey
	BencodingDictVal
	BencodingListItem
	BencodingPartial
)

func (e BencodingType) String() string {
	switch e {
	case BencodingASCIIString:
		return "BencodingASCIIString"
	case BencodingInt64:
		return "BencodingInt64"
	case BencodingListStart:
		return "BencodingListStart"
	case BencodingListEnd:
		return "BencodingListEnd"
	case BencodingListItem:
		return "BencodingListItem"
	case BencodingDictStart:
		return "BencodingDictStart"
	case BencodingDictKey:
		return "BencodingDictKey"
	case BencodingDictVal:
		return "BencodingDictVal"
	case BencodingDictEnd:
		return "BencodingDictEnd"
	case BencodingPartial:
		return "BencodingPartial"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}

type BencodingToken struct {
	Data []byte
	// index int
	Label BencodingType
}

func (obj BencodingToken) String() string {
	return string(obj.Data)
}

// ParseBencoding returns a collection of BencodingTokens parsed from a buffer.
func ParseBencoding(head int, data []byte, ch chan BencodingToken) (int, error) {
	var err error
	tail := head

	if len(data) < 2 {
		return head, fmt.Errorf("%v: %v", ErrBencoding, string(data))
	}

	switch {
	case data[head] >= '0' && data[head] <= '9':
		tail, err = newBencodingByteString(head, data, ch)
		if err != nil {
			return tail, err
		}
	case data[head] == 'i':
		tail, err = newBencodingInt64(head, data, ch)
		if err != nil {
			return tail, err
		}
	case data[head] == 'l':
		ch <- BencodingToken{data[head : head+1], BencodingListStart}
		var listClosed bool = false

		tail = head
		tail++

		for tail < len(data) {
			if data[tail] == 'e' {
				ch <- BencodingToken{data[head : tail+1], BencodingListEnd}
				listClosed = true
				break
			}

			// parse next list element
			ch <- BencodingToken{nil, BencodingListItem}
			tail, err = ParseBencoding(tail, data, ch)
			if err != nil {
				return tail, err
			}
			tail++
		}

		if !listClosed {
			return tail, fmt.Errorf("%v; %v (%v): '%v'", ErrBencodingList, ErrBencodingMissingToken, BencodingListEnd, string(data[head:tail+1]))
		}

	case data[head] == 'd':
		ch <- BencodingToken{data[head : head+1], BencodingDictStart}
		var dictClosed bool = false

		tail = head
		tail++

		// empty dict
		if data[tail] == 'e' {
			ch <- BencodingToken{data[head : tail+1], BencodingDictEnd}
			return tail, nil
		}

		for tail < len(data) {

			if data[tail] == 'e' {
				ch <- BencodingToken{data[head : tail+1], BencodingDictEnd}
				dictClosed = true
				break
			}

			// key token must be BencodingASCIIString
			if data[tail] < '0' || data[tail] > '9' {
				return tail, fmt.Errorf("%v: '%v'", ErrBencodingKeyInvalid, string(data[head:tail+1]))
			}
			// parse key token
			ch <- BencodingToken{nil, BencodingDictKey}
			tail, err = ParseBencoding(tail, data, ch)
			if err != nil {
				return tail, err
			}
			tail++

			// value token must exist for each key token
			if tail >= len(data) {
				return tail, fmt.Errorf("%v; %v : '%v'", ErrBencodingDict, ErrBencodingKeyNoValue, string(data[head:]))
			}

			// parse value token
			ch <- BencodingToken{nil, BencodingDictVal}
			tail, err = ParseBencoding(tail, data, ch)
			if err != nil {
				return tail, err
			}
			tail++

		}

		if !dictClosed {
			return tail, fmt.Errorf("%v; %v (%v): '%v'", ErrBencodingDict, ErrBencodingMissingToken, BencodingDictEnd, string(data[head:tail+1]))
		}

	default:
		return tail, fmt.Errorf("%v: '%v'", ErrBencoding, string(data[head:tail+1]))
	}

	return tail, err
}

// newBencodingByteString attempts to parse data for a valid ByteString Token from the provided buffer.
// Encoding: <string length encoded in base ten ASCII>:<string data>
// (https://wiki.theory.org/BitTorrentSpecification#Byte_Strings)
func newBencodingByteString(head int, data []byte, ch chan BencodingToken) (int, error) {
	// check the start of the ByteString is valid

	if len(data) < 2 {
		return 0, fmt.Errorf("BencodingByteStringType must be at least two characters")
	}

	switch {
	case data[head] == '0' && data[head+1] == ':':
		// 0-length ByteString represents an empty string
		ch <- BencodingToken{data[head : head+2], BencodingASCIIString}
		return head + 1, nil
	case data[head] >= '1' && data[head] <= '9':
		// noop (valid starting digits for non-empty strings)
	case data[head] == '0' && data[head] != ':':
		return head, fmt.Errorf("non-empty ByteString with leading 0 not allowed")
	default:
		return head, fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[0:1]))
	}

	// ensure all characters up-to ':' delimiter are only digits
	tail := head
	for tail < len(data) {
		tail++
		if data[tail] == ':' {
			break
		}
		if data[tail] < '0' || data[tail] > '9' {
			return tail, fmt.Errorf("only digits are allowed in the length portion of a ByteString (%v)", string(data[head:tail+1]))
		}
	}
	if data[tail] != ':' && tail == len(data)-1 {
		return tail, fmt.Errorf("exhaused input buffer but ByteString delimeter ':' never reached (%s)", string(data[head:tail+1]))
	}

	// determine expected length of string data
	lenData := string(data[head:tail])
	strLen, err := strconv.ParseUint(lenData, 10, 64)
	if err != nil {
		return tail, fmt.Errorf("invalid uint64 (%s): %v", lenData, err)
	}

	// shift tail to start of ASCII string data
	tail++

	// return token for Bencoding Byte String if sufficient data for indicated string length
	if availableData := (uint64(len(data)) - uint64(tail)); strLen > availableData {
		return tail, fmt.Errorf(
			"expecting more string data (%d) than is available (%d) in '%s'",
			strLen, availableData, string(data[tail:(uint64(tail)+availableData)]),
		)
	}
	// shift tail to end of ASCII string data segment
	// NOTE: uint64 -> int may lead to data loss
	tail += int(strLen) - 1
	// final token references slices for both whole token and data segments
	ch <- BencodingToken{data[head : tail+1], BencodingASCIIString}
	return tail, nil
}

// BencodingASCIIStringIntoString tries to return the string
// equivalent of the data in the provided buffer.
func BencodingASCIIStringIntoString(data []byte) (string, error) {
	for idx, c := range data {
		if c == ':' {
			strLen, err := strconv.ParseUint(string(data[:idx]), 10, 64)
			if err != nil {
				return "", err
			}
			if strLen+uint64(idx) < uint64(len(data)) {
				return string(data[idx+1:]), nil
			}
		}
	}
	return "", fmt.Errorf("missing ':' delimiter")
}

// newBencodingInt64 attempts to parse data for a valid Int64 Token from the provided buffer.
// Encoding: i<integer encoded in base ten ASCII>e
// (https://wiki.theory.org/BitTorrentSpecification#Integers)
func newBencodingInt64(head int, data []byte, ch chan BencodingToken) (int, error) {
	if len(data) < 3 {
		return head, fmt.Errorf("BencodingInt64 must be at least three characters")
	}

	// verify first character is int64 start delimiter
	if data[head] != 'i' {
		return head, fmt.Errorf("first character of bencodingInt64 must be 'i' (%v)", string(data))
	}

	// verify initial character combos are valid
	head++
	switch {
	case data[head] == '-' && (data[head+1] < '1' || data[head+1] > '9'):
		return head, fmt.Errorf("only digits 1-9 can immediately follow a minus symbol (%v)", string(data[head:head+2]))
	case data[head] == '0' && data[head+1] == 'e':
		ch <- BencodingToken{data[head : head+2], BencodingInt64} // zero-value bencodingInt64
		return head, nil
	case data[head] == '0' && data[head+1] != 'e':
		return head, fmt.Errorf("cannot have a leading '0' that is not immediately terminated with an 'e' (%v)", string(data[head:head+2]))
	case data[head] != '-' && (data[head] < '0' || data[head] > '9'):
		return head, fmt.Errorf("only digits 0-9 or minus symbol can start a BencodingInt64Type (%v)", string(data[head:head+1]))
	}

	// scan remainder of buffer
	tail := head
	tail++
	for tail < len(data) {
		switch {
		case data[tail] == 'e':
			_, err := strconv.ParseInt(string(data[head:tail]), 10, 64)
			if err != nil {
				return tail, fmt.Errorf("could not parse into int64 (%v)", string(data[head:tail]))
			}
			ch <- BencodingToken{data[head-1 : tail+1], BencodingInt64}
			return tail, nil
		case data[tail] < '0' || data[tail] > '9':
			return tail, fmt.Errorf("illegal character (%c) detected while parsing int64 bytes (%s)", data[tail], string(data[head:tail+1]))
		}
		tail++
	}
	ch <- BencodingToken{data[head : tail+1], BencodingPartial}
	return tail, fmt.Errorf("exhuasted buffer while parsing as bencodingInt64 (%s)", string(data[:tail+1]))
}

func BencodingInt64IntoInt64(data []byte) (int64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("%v: buffer is empty", ErrBencodingInt64)
	}
	if data[0] != 'i' {
		return 0, fmt.Errorf("%v: must start with 'i' delimiter (received: '%v')", ErrBencodingInt64, data[0])
	}
	for idx, c := range data[1:] {
		if c == 'e' {
			result, err := strconv.ParseInt(string(data[1:idx+1]), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("%v: %v", ErrBencodingInt64, err)
			}
			return result, nil
		}
	}
	return 0, fmt.Errorf("%v: missing 'e' delimiter", ErrBencodingInt64)
}
