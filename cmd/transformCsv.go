package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

// transformCsvCmd represents the csvTransform command
var transformCsvCmd = &cobra.Command{
	Use:   "csv",
	Short: "Upload a CSV to Google Sheets",
	Run:   transformCsv,
}

var folderId string

func init() {
	transformCmd.AddCommand(transformCsvCmd)
	transformCsvCmd.Flags().StringVar(&folderId, "folder", "", "The Google Sheet folder to upload your CSV to")
}

func transformCsv(cmd *cobra.Command, args []string) {
	_, err := os.Stat(source)
	if err != nil {
		slog.Info("Source file does not exist", "source", source, "err", err)
		os.Exit(1)
	}

	if folderId == "" {
		slog.Error("Folder flag is required")
		os.Exit(1)
	}
	ctx := context.Background()

	driveService, err := drive.NewService(ctx, option.WithScopes(drive.DriveScope))
	if err != nil {
		slog.Error("Unable to create Drive client", "err", err)
		os.Exit(1)
	}

	fileMetadata := &drive.File{
		Name:     fmt.Sprintf("ETD Ingest %s", time.Now().Format("2006-01-02")),
		MimeType: "application/vnd.google-apps.spreadsheet",
		Parents: []string{
			folderId,
		},
	}

	createdFile, err := driveService.Files.Create(fileMetadata).Do()
	if err != nil {
		slog.Error("Failed to create spreadsheet", "err", err)
		os.Exit(1)
	}

	spreadsheetID := createdFile.Id
	sheetsService, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		slog.Error("Unable to create Sheets client", "err", err)
		os.Exit(1)

	}

	data, err := readCSV(source)
	if err != nil {
		slog.Error("Failed to read CSV", "err", err)
		os.Exit(1)

	}

	err = writeToSheet(sheetsService, spreadsheetID, data)
	if err != nil {
		slog.Error("Failed to write to Google Sheet", "err", err)
		os.Exit(1)
	}

	fmt.Printf("https://docs.google.com/spreadsheets/d/%s", spreadsheetID)
}

func readCSV(filename string) ([][]interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rawData, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var data [][]interface{}
	for _, row := range rawData {
		var rowData []interface{}
		for _, col := range row {
			rowData = append(rowData, col)
		}
		data = append(data, rowData)
	}
	return data, nil
}

func writeToSheet(sheetsService *sheets.Service, spreadsheetID string, data [][]interface{}) error {
	writeRange := "Sheet1!A1"
	valueRange := &sheets.ValueRange{
		Values: data,
	}
	_, err := sheetsService.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).
		ValueInputOption("RAW").Do()

	return err
}
