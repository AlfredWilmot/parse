package internal

import (
	"testing"
)

func TestParseBencoding(t *testing.T) {
	cases := []struct {
		in, want string
		err      error
	}{
		{"0:", "0:", nil},
		{"3:foo", "3:foo", nil},
		{"4:foo", "", ErrBencodingByteString},
		{"3:fooo", "3:foo", nil},
	}
	for _, c := range cases {
		gotVal, gotErr := ParseIntoBencoding([]byte(c.in))
		if gotVal.String() != c.want {
			t.Errorf("ParseIntoBencoding(%v) == (%v, %v), want (%v, %v), ", c.in, gotVal.String(), gotErr, c.want, c.err)
		}
	}
}
