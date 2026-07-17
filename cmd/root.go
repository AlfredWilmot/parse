/*
Copyright © 2026 AlfredWilmot
*/
package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "parse",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		err := parseArgs(cmd, args)
		if err != nil {
			fmt.Println(err)
		}
	},
	Args: cobra.ExactArgs(2),
}

type DataFormat int32

const (
	Bencoding DataFormat = iota
	JSONFormat
)

var dataFormatMap = map[DataFormat]string{
	Bencoding:  "bencoding",
	JSONFormat: "json",
}

func (ss DataFormat) String() string {
	return dataFormatMap[ss]
}

// return the equivalent DataFormat corresponding to the input string, if one exists.
func getDataFormat(arg *string) (DataFormat, error) {
	switch *arg {
	case Bencoding.String():
		return Bencoding, nil
	case JSONFormat.String():
		return JSONFormat, nil
	default:
		return 0, fmt.Errorf("unrecognised DataFormat '%v'", *arg)
	}
}

func parseArgs(cmd *cobra.Command, args []string) error {

	inputDataformat, err := getDataFormat(&args[0])
	if err != nil {
		return err
	}
	outputDataformat, err := getDataFormat(&args[1])
	if err != nil {
		return err
	}

	fmt.Printf("Transforming from %v to %v\n", inputDataformat, outputDataformat)

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.parse.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
