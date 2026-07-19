/* Package internal */
package internal

import (
	"bytes"
	"os"
)

func BencodingToJSON(buffer *os.File) {
	// TODO
}

func JSONToBencoding(buffer *os.File) {
	// TODO
}

func FillBufferFromStdin(buffer *bytes.Buffer) {
	// fill buffer with contents of stdin
	for {
		n, err := buffer.ReadFrom(os.Stdin)
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
		if n <= 0 {
			break
		}
	}
}
