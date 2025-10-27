/*
Copyright © 2025 Isaac Striker isaacstriker@pm.me

*/
package cmd

import (
	"os"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagMode int
	flagHeight string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chell <file.wav>",
	Short: "Edit small waveforms in the terminal",
	Long: `Chell is an audio editing tool aimed towards creating small samples for 
use in your favorite DAW. Supports '.wav' formats, as well as:

- Chopping and cutting waveform.
- Cutting and swapping waveform data.

If you have an idea for a feature, feel free to contribute, or fork the repo!`,
	Args: cobra.ExactArgs(1),
	RunE: func (cmd *cobra.Command, args []string) error {
		path := args[0]
		return runFile(path, flagHeight, flagMode, os.Stdout)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.Flags().StringVarP(&flagHeight, "height", "H", "braille", "vertical character rows used for rendering")
	rootCmd.Flags().IntVarP(&flagMode, "mode", "m", 12, "render mode: braille|ascii")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chell.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


