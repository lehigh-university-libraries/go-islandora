package cmd

import (
	"github.com/spf13/cobra"
)

var baseUrl string

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export content",
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.PersistentFlags().StringVar(&baseUrl, "baseUrl", "", "The base URL to export from (e.g. https://google.com)")
}
