/*
Package parse Copyright © 2026 AlfredWilmot
*/
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"parse/internal"

	"github.com/spf13/cobra"
)

// used for flags (see init() for details)
var (
	raw bool
)

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
			internal.BencodingToJSON(os.Stdin)
		} else {
			os.Exit(1)
		}
	case inputDataFormat == JSON && outputDataFormat == BENCODING:
		if raw {
			internal.JSONToBencoding(os.Stdin)
		} else {
			os.Exit(1)
		}
	case inputDataFormat == outputDataFormat:
		// write contents of filled buffer to stdout
		internal.FillBufferFromStdin(&buffer)
		buffer := bytes.TrimRight(buffer, "\x00")
		result := strings.TrimRight(string(buffer), "\n")
		fmt.Println(result)
	}
	fmt.Printf("Converting from '%v' to '%v'\n", args[0], args[1])
}

func init() {
	rootCmd.Flags().BoolVarP(&raw, "raw", "r", false, "Directly parse from input to output without populating an intermediate data-structure.")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
