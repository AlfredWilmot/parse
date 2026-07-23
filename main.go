/*
Package parse Copyright © 2026 AlfredWilmot
*/
package main

import (
	"fmt"
	"os"

	"parse/internal"

	"github.com/spf13/cobra"
)

// used for flags (see init() for details)
var (
	stream bool
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

	// TODO: handle errors when reading from input
	inputBuffer := make([]byte, 1<<30) // 1GB
	n, _ := os.Stdin.Read(inputBuffer)
	inputBuffer = inputBuffer[:n]
	fmt.Println("buffer-len:", len(inputBuffer))

	// parse data into designated input type
	var tokenFeed chan internal.Tokener
	switch inputDataFormat {
	case BENCODING:
		var err error
		tokenFeed, err = internal.ParseBencoding(inputBuffer, 100)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case JSON:
		// TODO
	}

	// transform parsed data into designated output type

	switch {
	case inputDataFormat == outputDataFormat:
		// noop, just write inputBuffer directly to output
	case inputDataFormat == BENCODING && outputDataFormat == JSON:
		// TODO
	case inputDataFormat == JSON && outputDataFormat == BENCODING:
		// TODO
	}

	fmt.Printf("Converting from '%v' to '%v'\n", args[0], args[1])
	for token := range tokenFeed {
		fmt.Println(token)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&stream, "stream", false, "Stream contents of input through parser")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
