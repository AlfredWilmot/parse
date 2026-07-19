package internal

import (
	"fmt"
	"strings"
)

type bencodingType int

const (
	bencodingBytesType bencodingType = iota
	bencodingIntegerType
	bencodingListType
	bencodingDictionaryType
)

type bencoding interface {
	Type() bencodingType
	String() string
}

func bencodingGather(b bencoding, c chan string) {
}

type bencodingBytes struct {
	len int64 `validate:"gte=0"`
	buf []byte
}

func (b bencodingBytes) String() string {
	return fmt.Sprintf("%d:%s", b.len, string(b.buf))
}
func (b bencodingBytes) Type() bencodingType {
	return bencodingBytesType
}

type bencodingInteger int64

func (b bencodingInteger) String() string {
	return fmt.Sprintf("i%de", b)
}
func (b bencodingInteger) Type() bencodingType {
	return bencodingIntegerType
}

type bencodingList []bencoding

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
