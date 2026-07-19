/* Package internal */
package internal

import (
	"errors"
	"io"
	"os"
)

func BencodingToJSON(buffer *os.File) {
	// TODO
}

func JSONToBencoding(buffer *os.File) {
	// TODO
}

func FillBufferFromStdin(buffer *[]byte) {
	// fill buffer with contents of stdin
	for {
		n, err := os.Stdin.Read(*buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				os.Stderr.WriteString(err.Error())
				os.Exit(1)
			}
		}
		if n <= 0 {
			break
		}
	}
}
