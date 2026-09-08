package internal

import (
	"testing"
)

func TestParseBencodingByteString(t *testing.T) {
	cases := []struct {
		input  []byte
		expect string
	}{
		// success cases
		{[]byte("0:"), ""},
		{[]byte("3:foo"), "foo"},
		{[]byte("3:fooo"), "foo"},
		// error cases
		{[]byte("4:foo"), ""},
		{[]byte("04:foo"), ""},
		{[]byte("d0h:foo"), ""},
	}

	ch := make(chan BencodingToken)

	go func() {
		defer close(ch)
		for _, c := range cases {
			newBencodingByteString(0, []byte(c.input), ch)
		}
	}()

	for _, c := range cases {
		gotVal, _ := BencodingASCIIStringIntoString((<-ch).Data)
		if gotVal != c.expect {
			t.Errorf("newBencodingByteString(%v) == (%v), want (%v), ", c.input, gotVal, c.expect)
		}
	}
}

func TestParseBencodingInt64(t *testing.T) {
	cases := []struct {
		input  []byte
		expect int64
	}{
		// success cases
		{[]byte("i0e"), 0},
		{[]byte("i-1e"), -1},
		// error cases
		{[]byte("a2e"), 0},
		{[]byte("i05-1e"), 0},
		{[]byte("i00e"), 0},
		{[]byte("ie"), 0},
	}

	ch := make(chan BencodingToken)

	go func() {
		defer close(ch)
		for _, c := range cases {
			newBencodingInt64(0, []byte(c.input), ch)
		}
	}()

	for _, c := range cases {
		gotVal, _ := BencodingInt64IntoInt64((<-ch).Data)
		if gotVal != c.expect {
			t.Errorf("newBencodingInt64(%v) == (%v), want (%v), ", c.input, gotVal, c.input)
		}
	}
}
