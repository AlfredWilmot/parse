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
		if string(gotVal.ParsedData()) != string(c.want) {
			t.Errorf("newBencodingByteString(%v) == (%v), want (%v), ", c.in, string(gotVal.ParsedData()), string(c.want))
		}
	}
}

func TestParseBencodingInt64(t *testing.T) {
	cases := []struct {
		in        string
		wantSlice []byte
	}{
		{"i0e", []byte("0")},
		{"i-1e", []byte("-1")},
		{"a2e", nil},
		{"i05-1e", nil},
		{"i00e", nil},
		{"ie", nil},
	}
	for _, c := range cases {
		gotVal, _ := newBencodingInt64([]byte(c.in))
		if string(gotVal.ParsedData()) != string(c.wantSlice) {
			t.Errorf("newBencodingInt64(%v) == (%v), want (%v), ", c.in, string(gotVal.ParsedData()), string(c.wantSlice))
		}
	}
}
