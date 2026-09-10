package cmd

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lehigh-university-libraries/go-islandora/pkg/proquest"
	"github.com/spf13/cobra"
	"golang.org/x/net/html/charset"
)

var etdEmbargoFutureOnly bool

var transformEtdEmbargoCmd = &cobra.Command{
	Use:   "etd-embargo",
	Short: "Extract repository embargo dates and local restrictions from ETD XML files",
	Long: `Recursively read XML files under --source and write a CSV containing
file, field_edtf_date_embargo, field_local_restriction, and pdf_file.
Output goes to stdout, or to --target when supplied.`,
	Args: cobra.NoArgs,
	RunE: extractETDEmbargoes,
}

func init() {
	transformCmd.AddCommand(transformEtdEmbargoCmd)
	transformEtdEmbargoCmd.Flags().BoolVar(&etdEmbargoFutureOnly, "future-only", false, "Only include embargo dates after today (UTC)")
}

func extractETDEmbargoes(cmd *cobra.Command, args []string) error {
	today := time.Now().UTC().Format("2006-01-02")
	if isDir, err := isDirectory(source); err != nil || !isDir {
		return fmt.Errorf("source flag is not a directory: %s", source)
	}
	out := cmd.OutOrStdout()
	if target != "" {
		file, err := os.Create(target)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		out = file
	}
	writer := csv.NewWriter(out)
	defer writer.Flush()
	if err := writer.Write([]string{"file", "field_edtf_date_embargo", "field_local_restriction", "pdf_file"}); err != nil {
		return err
	}
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".xml") {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		decoder := xml.NewDecoder(file)
		decoder.CharsetReader = charset.NewReaderLabel
		var submission proquest.DISSSubmission
		if err := decoder.Decode(&submission); err != nil {
			return fmt.Errorf("failed to decode XML %s: %w", path, err)
		}
		embargoDate, err := submission.EmbargoDate()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		// EmbargoDate returns YYYY-MM-DD, so lexical order matches date order.
		if etdEmbargoFutureOnly && embargoDate <= today {
			return nil
		}
		return writer.Write([]string{path, embargoDate, strconv.FormatBool(submission.Repository.LocalRestriction()), strings.TrimSpace(submission.Content.Binary.FileName)})
	})
	if err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}
