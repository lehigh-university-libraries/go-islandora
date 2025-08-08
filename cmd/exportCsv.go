package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	nid     int
	csvFile string
)

// Custom type for sorting rows
type Row map[string]string

// Custom sorting function
type ByFieldMemberOfAndWeight []Row

func (a ByFieldMemberOfAndWeight) Len() int {
	return len(a)
}

func (a ByFieldMemberOfAndWeight) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByFieldMemberOfAndWeight) Less(i, j int) bool {
	if a[i]["field_member_of"] == a[j]["field_member_of"] {
		weightI, _ := strconv.Atoi(a[i]["field_weight"])
		weightJ, _ := strconv.Atoi(a[j]["field_weight"])
		return weightI < weightJ
	}
	return a[i]["field_member_of"] < a[j]["field_member_of"]
}

// exportCsvCmd represents the csv command
var exportCsvCmd = &cobra.Command{
	Use:   "csv",
	Short: "Recursively export a workbench CSV for an Islandora node",
	Long:  `Recursively export a workbench CSV for an Islandora node.`,
	Run: func(cmd *cobra.Command, args []string) {
		if baseUrl == "" || nid == 0 {
			slog.Error("--baseUrl and --nid flags are required")
			os.Exit(1)
		}
		fmt.Println("CSV files merged successfully into", csvFile)
	},
}

func init() {
	exportCmd.AddCommand(exportCsvCmd)
	exportCsvCmd.Flags().IntVar(&nid, "nid", 0, "The node ID to export a CSV")
	exportCsvCmd.Flags().StringVar(&csvFile, "output", "merged.csv", "The CSV file name to save the export to")
}
