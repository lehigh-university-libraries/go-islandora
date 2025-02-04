/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// transformCsvCmd represents the csvTransform command
var transformCsvCmd = &cobra.Command{
	Use:   "csv",
	Short: "Transform CSV to another format",
}

func init() {
	transformCmd.AddCommand(transformCsvCmd)
}
