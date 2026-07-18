// Package cmd does a thing
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-playground/validator"
	"github.com/spf13/cobra"
)

const (
	BENCODING = "bencoding"
	JSON      = "json"
)

// used for flags
var (
	raw bool
)

var rootCmd = &cobra.Command{
	Use:   "parse [flags] input-format output-format",
	Short: "Transform from one data-format to another",
	Long: `Transform from one data-format to another
Valid formats: [bencoding|json]`,
	Run:       runCmd,
	Args:      cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	ValidArgs: []string{"bencoding", "json"},
}

func runCmd(cmd *cobra.Command, args []string) {
	inputDataFormat := args[0]
	outputDataFormat := args[1]

	buffer := make([]byte, 1024)

	switch {
	case inputDataFormat == BENCODING && outputDataFormat == JSON:
		if raw {
			bencodingToJSON(os.Stdin)
		} else {
			os.Exit(1)
		}
	case inputDataFormat == JSON && outputDataFormat == BENCODING:
		if raw {
			jsonToBencoding(os.Stdin)
		} else {
			os.Exit(1)
		}
	case inputDataFormat == outputDataFormat:
		// fill buffer with contents of stdin
		for {
			n, err := os.Stdin.Read(buffer)
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
		// write contents of filled buffer to stdout
		fmt.Println(string(buffer))
	}
	fmt.Printf("Converting from '%v' to '%v'\n", args[0], args[1])
}

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

func bencodingToJSON(buffer *os.File) {
	// TODO
}

func jsonToBencoding(buffer *os.File) {
	// TODO
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&raw, "raw", "r", false, "Directly parse from input to output without populating an intermediate data-structure.")
}
