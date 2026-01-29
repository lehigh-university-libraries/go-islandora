package cmd

import (
	"archive/zip"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"log/slog"

	"github.com/lehigh-university-libraries/go-islandora/pkg/proquest"
	"github.com/spf13/cobra"
	"golang.org/x/net/html/charset"
)

var etdCsvFile string

// transformEtdDateBackfillCmd represents the etd-date-backfill command
var transformEtdDateBackfillCmd = &cobra.Command{
	Use:   "etd-date-backfill",
	Short: "Check ETD dates against CSV export and report mismatches",
	Long: `Scan ZIP files, extract title and DISS_accept_date from XML,
match against a CSV export, and report any mismatches where
field_edtf_date_issued_value doesn't match DISS_accept_date.

The CSV should have columns: title, nid, field_edtf_date_issued_value, field_edtf_date_embargo_value`,
	Run: etdDateBackfill,
}

func init() {
	transformCmd.AddCommand(transformEtdDateBackfillCmd)
	transformEtdDateBackfillCmd.Flags().StringVar(&etdCsvFile, "csv", "", "CSV export file with title, nid, field_edtf_date_issued_value, field_edtf_date_embargo_value columns")
	err := transformEtdDateBackfillCmd.MarkFlagRequired("csv")
	if err != nil {
		slog.Error("Unable to mark csv flag as required for etd-date-backfill command")
		os.Exit(1)
	}
}

// etdRecord represents a row from the CSV export
type etdRecord struct {
	nid                       string
	fieldEdtfDateIssuedValue  string
	fieldEdtfDateEmbargoValue string
}

func etdDateBackfill(cmd *cobra.Command, args []string) {
	isDir, err := isDirectory(source)
	if !isDir || err != nil {
		slog.Error("Source flag is not a directory", "source", source)
		os.Exit(1)
	}

	if etdCsvFile == "" {
		slog.Error("CSV flag is required")
		os.Exit(1)
	}

	// Read CSV into map keyed by title
	records, err := readCSVExport(etdCsvFile)
	if err != nil {
		slog.Error("Failed to read CSV export", "error", err)
		os.Exit(1)
	}
	slog.Info("Loaded CSV export", "records", len(records))

	// Print header for output
	fmt.Println("nid\tfield_edtf_date_issued_value\tfield_edtf_date_embargo_value")

	// Iterate over ZIP files
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Error("Error accessing file", "file", path, "error", err)
			return nil
		}
		if strings.HasSuffix(info.Name(), ".zip") {
			if err := processZipForDateCheck(path, records); err != nil {
				slog.Error("Failed to process ZIP", "file", path, "error", err)
			}
		}
		return nil
	})

	if err != nil {
		slog.Error("Failed to walk directory", "error", err)
	}
}

// readCSVExport reads the CSV export file and returns a map keyed by title
func readCSVExport(csvPath string) (map[string]etdRecord, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Find column indices
	titleIdx := -1
	nidIdx := -1
	dateIssuedIdx := -1
	dateEmbargoIdx := -1

	for i, col := range header {
		switch strings.TrimSpace(col) {
		case "title":
			titleIdx = i
		case "nid":
			nidIdx = i
		case "field_edtf_date_issued_value":
			dateIssuedIdx = i
		case "field_edtf_date_embargo_value":
			dateEmbargoIdx = i
		}
	}

	if titleIdx == -1 || nidIdx == -1 || dateIssuedIdx == -1 {
		return nil, fmt.Errorf("CSV must have columns: title, nid, field_edtf_date_issued_value")
	}

	records := make(map[string]etdRecord)
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		title := ""
		if titleIdx < len(row) {
			title = strings.TrimSpace(row[titleIdx])
		}
		if title == "" {
			continue
		}

		record := etdRecord{}
		if nidIdx < len(row) {
			record.nid = strings.TrimSpace(row[nidIdx])
		}
		if dateIssuedIdx < len(row) {
			record.fieldEdtfDateIssuedValue = strings.TrimSpace(row[dateIssuedIdx])
		}
		if dateEmbargoIdx != -1 && dateEmbargoIdx < len(row) {
			record.fieldEdtfDateEmbargoValue = strings.TrimSpace(row[dateEmbargoIdx])
		}

		records[title] = record
	}

	return records, nil
}

// processZipForDateCheck extracts XML from a ZIP and checks dates against CSV
func processZipForDateCheck(zipPath string, records map[string]etdRecord) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	var xmlFile *zip.File
	for _, file := range r.File {
		if strings.HasSuffix(file.Name, "_DATA.xml") {
			xmlFile = file
			break
		}
	}

	if xmlFile == nil {
		return fmt.Errorf("no _DATA.xml file found in %s", zipPath)
	}

	xmlReader, err := xmlFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open XML file in ZIP: %w", err)
	}
	defer xmlReader.Close()

	decoder := xml.NewDecoder(xmlReader)
	decoder.CharsetReader = charset.NewReaderLabel

	var submission proquest.DISSSubmission
	if err := decoder.Decode(&submission); err != nil {
		return fmt.Errorf("failed to decode XML: %w", err)
	}

	title := strings.TrimSpace(submission.Description.Title)
	// field_edtf_date_issued_value comes from DISS_comp_date (extract just the year)
	completionDate := submission.Description.Dates.CompletionDate
	completionYear := strings.Split(completionDate, "-")[0]
	// embargo is calculated from DISS_accept_date
	embargoDate := submission.EmbargoDate()

	// Look up by full title first, then truncated title
	record, found := records[title]
	if !found && len(title) > 255 {
		record, found = records[title[0:255]]
	}

	if !found {
		slog.Warn("No matching title in CSV", "title", title, "zip", zipPath)
		return nil
	}

	// Check if dates match
	// The completionYear is just the year from DISS_comp_date
	// The field_edtf_date_issued_value should match
	if record.fieldEdtfDateIssuedValue != completionYear || record.fieldEdtfDateEmbargoValue != embargoDate {
		fmt.Printf("%s\t%s\t%s\n", record.nid, completionYear, embargoDate)
	}

	return nil
}
