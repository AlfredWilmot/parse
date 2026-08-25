/*
Package parse Copyright © 2026 AlfredWilmot
*/
package main

import (
	"fmt"
	"log/slog"
	"os"

	"parse/internal"

	"github.com/spf13/cobra"
)

// used for flags (see init() for details)
var (
	tokenise bool
	data     string
	quiet    bool
)

func init() {
	rootCmd.Flags().BoolVarP(&tokenise, "tokenise", "t", false, "Parse data into the tokens corresponding to the selected data-format")
	rootCmd.Flags().StringVarP(&data, "data", "d", "", "Pass data directly")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Silence logs")
}

const (
	BENCODING = "bencoding"
	JSON      = "json"
)

var rootCmd = &cobra.Command{
	Use:   "parse [flags] input-format output-format",
	Short: "Transform from one data-format to another",
	Long: `Transform from one data-format to another
Valid formats: [bencoding|json]`,
	Run:       runCmd,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{"bencoding", "json"},
}

const (
	MaxBufferLen  = 1 << 30 // 1GB
	MaxTokenCount = 100
)

var ErrBufferLen = fmt.Errorf("exceeded MaxBufferLen %d", MaxBufferLen)

func runCmd(cmd *cobra.Command, args []string) {
	// parse cli flags
	data, _ := cmd.Flags().GetString("data")
	tokenise, _ := cmd.Flags().GetBool("tokenise")
	quiet, _ := cmd.Flags().GetBool("quiet")

	// configure logging
	var logger *slog.Logger
	if quiet {
		logger = slog.New(slog.DiscardHandler)
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	}
	slog.SetDefault(logger)

	// determine desired dataformat
	inputDataFormat := args[0]

	// TODO: handle errors when reading-in data
	inputBuffer := make([]byte, MaxBufferLen)
	if len(data) > 0 {
		slog.Debug("recieved fixed data", "flag", "--data", "content", data)
		inputBuffer = []byte(data)
		if len(data) > MaxBufferLen {
			logger.Error(fmt.Sprintln(ErrBufferLen))
			os.Exit(1)
		}
	} else {
		slog.Info("streaming data", "source", "/dev/stdin")
		n, _ := os.Stdin.Read(inputBuffer)
		inputBuffer = inputBuffer[:n]
	}

	ch := make(chan internal.BencodingToken)
	var err error

	// parse data from designated input type
	switch inputDataFormat {
	case BENCODING:
		go func() {
			defer close(ch)
			_, err = internal.ParseBencoding(0, inputBuffer, ch)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}()

	case JSON:
		// TODO
	}

	// used to handle qualifier for the current label
	prevLabel := internal.BencodingNull

	for v := range ch {

		if tokenise {
			fmt.Fprintf(os.Stderr, "%v", v.Label)
			if v.Data != nil {
				fmt.Fprintf(os.Stderr, " --> (%s)\n", v)
			} else {
				fmt.Fprint(os.Stderr, "\n")
			}
		}

		switch v.Label {
		case internal.BencodingDictStart:
			fmt.Print("{")
		case internal.BencodingDictEnd:
			fmt.Print("}")
		case internal.BencodingListStart:
			fmt.Print("[")
		case internal.BencodingListEnd:
			fmt.Print("]")
		case internal.BencodingInt64:
			result, err := internal.BencodingInt64IntoInt64(v.Data)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("%d", result)

		case internal.BencodingASCIIString:
			result, err := internal.BencodingASCIIStringIntoString(v.Data)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("\"%s\"", result)
			if prevLabel == internal.BencodingDictKey {
				fmt.Printf(": ")
			}
		case internal.BencodingDictKey:
			if prevLabel != internal.BencodingDictStart {
				fmt.Print(", ")
			}
		case internal.BencodingListItem:
			if prevLabel != internal.BencodingListStart {
				fmt.Print(", ")
			}

		}

		prevLabel = v.Label
	}
	fmt.Print("\n")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
