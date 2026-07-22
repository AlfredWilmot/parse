package internal

import (
	"testing"
)

func TestParseBencodingByteString(t *testing.T) {
	cases := []struct {
		in   string
		want []byte
	}{
		{"0:", []byte("")},
		{"3:foo", []byte("foo")},
		{"3:fooo", []byte("foo")},
		{"4:foo", nil},
		{"04:foo", nil},
		{"d0h:foo", nil},
	}
	for _, c := range cases {
		gotVal, _ := newBencodingByteString([]byte(c.in))
		if string(gotVal) != string(c.want) {
			t.Errorf("ParseIntoBencoding(%v) == (%v), want (%v), ", c.in, string(gotVal), string(c.want))
		}
	}
}
