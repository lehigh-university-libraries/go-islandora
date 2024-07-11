package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var nid, csvFile string
var parentModels = []string{
	"Collection",
	"Compound Object",
	"Paged Content",
	"Publication Issue",
	"Sub-Collection",
}

// csvCmd represents the csv command
var csvCmd = &cobra.Command{
	Use:   "csv",
	Short: "Recursively export a workbench CSV for an Islandora node",
	Long:  `Recursively export a workbench CSV for an Islandora node.`,
	Run: func(cmd *cobra.Command, args []string) {

		if baseUrl == "" || nid == "" {
			slog.Error("--baseUrl and --nid flags are required")
			os.Exit(1)
		}

		baseURL := fmt.Sprintf("%s/node/%s?_format=workbench_csv", baseUrl, nid)

		var allHeaders []string
		headerMap := make(map[string]bool)
		rows := []map[string]string{}
		nodeIDMap := make(map[string]bool)

		// Fetch the initial CSV
		initialCSV, err := fetchCSV(baseURL)
		if err != nil {
			log.Fatal(err)
		}

		// Process the initial CSV to find unique columns and rows to fetch
		for _, record := range initialCSV[1:] { // Skip header row
			row := make(map[string]string)
			nodeID := ""

			for i, header := range initialCSV[0] {
				if header == "node_id" {
					nodeID = record[i]
				}
				if !headerMap[header] {
					allHeaders = append(allHeaders, header)
					headerMap[header] = true
				}
				row[header] = record[i]
			}

			if !nodeIDMap[nodeID] {
				rows = append(rows, row)
				nodeIDMap[nodeID] = true
			}

			if StrInSlice(row["field_model"], parentModels) {
				subNodeID := row["node_id"]
				subURL := fmt.Sprintf("%s/node/%s?_format=workbench_csv", baseUrl, subNodeID)
				subCSV, err := fetchCSV(subURL)
				if err != nil {
					slog.Error("Failed to fetch sub-collection CSV for node ID", "nid", subNodeID, "err", err)
					continue
				}

				for _, subRecord := range subCSV[1:] { // Skip header row
					subRow := make(map[string]string)
					subNodeID := ""

					for i, subHeader := range subCSV[0] {
						if subHeader == "node_id" {
							subNodeID = subRecord[i]
						}
						if !headerMap[subHeader] {
							allHeaders = append(allHeaders, subHeader)
							headerMap[subHeader] = true
						}
						subRow[subHeader] = subRecord[i]
					}

					if !nodeIDMap[subNodeID] {
						rows = append(rows, subRow)
						nodeIDMap[subNodeID] = true
					}
				}
			}
		}

		// Write to the output CSV
		outFile, err := os.Create(csvFile)
		if err != nil {
			log.Fatal(err)
		}
		defer outFile.Close()

		csvWriter := csv.NewWriter(outFile)
		defer csvWriter.Flush()

		csvWriter.Write(allHeaders)

		for _, row := range rows {
			record := make([]string, len(allHeaders))
			for i, header := range allHeaders {
				record[i] = row[header]
			}
			csvWriter.Write(record)
		}

		fmt.Println("CSV files merged successfully into", csvFile)
	},
}

func init() {
	exportCmd.AddCommand(csvCmd)
	csvCmd.Flags().StringVar(&nid, "nid", "", "The node ID to export a CSV")
	csvCmd.Flags().StringVar(&csvFile, "output", "merged.csv", "The CSV file name to save the export to")
}

func fetchCSV(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch CSV from %s: %s", url, resp.Status)
	}

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func StrInSlice(s string, sl []string) bool {
	for _, a := range sl {
		if a == s {
			return true
		}
	}
	return false
}
