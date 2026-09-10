package cmd

import (
	"archive/zip"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
	Short: "Check ETD dates and local restrictions against a CSV export",
	Long: `Scan ZIP files and report date or local-restriction mismatches as TSV.
Match records by full or truncated title.

The CSV must include title, nid, and field_edtf_date_issued_value.
Optional columns: field_edtf_date_embargo_value and field_local_restriction.`,
	RunE: etdDateBackfill,
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
	localRestriction          bool
}

func etdDateBackfill(cmd *cobra.Command, args []string) error {
	isDir, err := isDirectory(source)
	if !isDir || err != nil {
		return fmt.Errorf("source flag is not a directory: %s", source)
	}

	if etdCsvFile == "" {
		return fmt.Errorf("CSV flag is required")
	}

	// Index CSV records by title.
	records, err := readCSVExport(etdCsvFile)
	if err != nil {
		return fmt.Errorf("failed to read CSV export: %w", err)
	}
	slog.Info("Loaded CSV export", "records", len(records))

	// Print header for output
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), "nid\tfield_edtf_date_issued_value\tfield_edtf_date_embargo_value\tfield_local_restriction"); err != nil {
		return err
	}

	// Iterate over ZIP files
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".zip") {
			if err := processZipForDateCheck(path, records, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("failed to process ZIP %s: %w", path, err)
			}
		}
		return nil
	})

	return err
}

// readCSVExport indexes records by title.
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
	localRestrictionIdx := -1

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
		case "field_local_restriction":
			localRestrictionIdx = i
		}
	}

	if titleIdx == -1 || nidIdx == -1 || dateIssuedIdx == -1 {
		return nil, fmt.Errorf("CSV must have columns: title, nid, field_edtf_date_issued_value")
	}

	records := make(map[string]etdRecord)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
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
		if localRestrictionIdx != -1 && localRestrictionIdx < len(row) {
			if value := strings.TrimSpace(row[localRestrictionIdx]); value != "" {
				record.localRestriction, err = strconv.ParseBool(value)
				if err != nil {
					return nil, fmt.Errorf("invalid field_local_restriction for nid %s: %w", record.nid, err)
				}
			}
		}

		records[title] = record
	}

	return records, nil
}

// processZipForDateCheck extracts XML from a ZIP and checks dates against CSV
func processZipForDateCheck(zipPath string, records map[string]etdRecord, out io.Writer) error {
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
	embargoDate, err := submission.EmbargoDate()
	if err != nil {
		return err
	}

	// Look up by full title first, then truncated title.
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
	localRestriction := submission.Repository.LocalRestriction()
	if record.fieldEdtfDateIssuedValue != completionYear || record.fieldEdtfDateEmbargoValue != embargoDate || record.localRestriction != localRestriction {
		_, err := fmt.Fprintf(out, "%s\t%s\t%s\t%t\n", record.nid, completionYear, embargoDate, localRestriction)
		return err
	}

	return nil
}
