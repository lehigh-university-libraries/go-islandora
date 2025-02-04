package cmd

import (
	"github.com/spf13/cobra"
)

var target string
var source string

// transformCmd represents the transform command
var transformCmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform one storage format to another",
}

func init() {
	rootCmd.AddCommand(transformCmd)
	transformCmd.PersistentFlags().StringVar(&source, "source", "", "Source file")
	transformCmd.PersistentFlags().StringVar(&target, "target", "", "Where to save target file")
}
