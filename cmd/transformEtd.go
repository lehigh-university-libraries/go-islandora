package cmd

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/lehigh-university-libraries/go-islandora/pkg/islandora"
	"github.com/lehigh-university-libraries/go-islandora/pkg/proquest"
	"github.com/spf13/cobra"
	"golang.org/x/net/html/charset"
)

// transformEtdCmd represents the csvTransform command
var transformEtdCmd = &cobra.Command{
	Use:   "etd",
	Short: "Transform a directory of ProQuest ETD ZIPs to another format",
	RunE:  transformETDs,
}

func init() {
	transformCmd.AddCommand(transformEtdCmd)
}

// transformETDs reads ZIP files from "etds" directory, extracts XML, and writes a CSV
func transformETDs(cmd *cobra.Command, args []string) error {
	isDir, err := isDirectory(source)
	if !isDir || err != nil {
		return fmt.Errorf("source flag is not a directory: %s", source)
	}

	if target == "" {
		return fmt.Errorf("target flag is required")
	}

	outFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	header := []string{
		"Upload ID",
		"Page/Item Parent ID",
		"Child Sort Order",
		"Node ID",
		"Parent Collection",
		"Object Model",
		"File Path",
		"Add Coverpage (Y/N)",
		"Title",
		"Full Title",
		"Make Public (Y/N)",
		"Contributor",
		"Related Department",
		"Resource Type",
		"Genre (Getty AAT)",
		"Creation Date",
		"Embargo Until Date",
		"Language",
		"File Format (MIME Type)",
		"Digital Origin",
		"Abstract",
		"Subject Topic (LCSH)",
		"Keyword",
		"Rights Statement",
		"Supplemental File",
		"Local Restriction",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Iterate over ZIP files
	uploadId := 1
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".zip") {
			slog.Info("Processing ZIP file", "file", path)
			if err := processZip(uploadId, path, writer); err != nil {
				return fmt.Errorf("failed to process ZIP %s: %w", path, err)
			}
			uploadId += 1
		}
		return nil
	})

	if err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

// processZip extracts XML from a ZIP, finds the relevant data, and writes to CSV
func processZip(uploadId int, zipPath string, writer *csv.Writer) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	var xmlFile *zip.File
	var pdfFile string
	var supplementaryFiles []string

	for _, file := range r.File {
		if strings.HasSuffix(file.Name, "_DATA.xml") {
			xmlFile = file
		} else if strings.HasSuffix(file.Name, ".pdf") {
			pdfFile = filepath.Join(source, file.Name)
		} else if strings.Contains(file.Name, "/") && !strings.HasSuffix(file.Name, "/") {
			supplementaryFiles = append(supplementaryFiles, filepath.Join(source, file.Name))
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
	decoder.CharsetReader = charset.NewReaderLabel // Enable charset conversion

	var submission proquest.DISSSubmission
	if err := decoder.Decode(&submission); err != nil {
		return fmt.Errorf("failed to decode XML: %w", err)
	}
	embargoDate, err := submission.EmbargoDate()
	if err != nil {
		return err
	}

	// Format keywords
	var keywords []string
	for _, keyword := range submission.Description.Categorization.Keywords {
		keyword = strings.ReplaceAll(keyword, ", ", " ; ")
		keywords = append(keywords, keyword)
	}

	// Format subjects
	var subjects []string
	for _, category := range submission.Description.Categorization.Categories {
		subjects = append(subjects, category.Description)
	}

	completionYear, err := parseCompletionYear(submission.Description.Dates.CompletionDate)
	if err != nil {
		slog.Error("Invalid completion year format", "date", submission.Description.Dates.CompletionDate, "error", err)
		return err
	}
	title := submission.Description.Title
	if len(title) > 255 {
		title = submission.Description.Title[0:255]
	}
	genre := "theses"
	if submission.Description.Degree == "Ph.D." {
		genre = "dissertations"
	}
	language := submission.Description.Categorization.Language
	if language == "en" {
		language = "English"
	} else {
		return fmt.Errorf("unknown language: %s", language)
	}
	row := []string{
		strconv.Itoa(uploadId),
		"",
		"",
		"",
		"202",
		"Digital Document",
		pdfFile,
		"Yes",
		title,
		submission.Description.Title,
		"Yes",
		getContributors(submission),
		submission.Description.Institution.Department,
		"Text",
		genre,
		completionYear,
		embargoDate,
		language,
		"application/pdf",
		"born digital",
		"<p>" + strings.Join(submission.Content.Abstract.Paragraphs, "</p><p>") + "</p>",
		strings.Join(subjects, " ; "),
		strings.Join(keywords, " ; "),
		"IN COPYRIGHT",
		strings.Join(supplementaryFiles, " ; "),
		strconv.FormatBool(submission.Repository.LocalRestriction()),
	}

	// Write row to CSV
	if err := writer.Write(row); err != nil {
		return fmt.Errorf("failed to write row to CSV: %w", err)
	}
	writer.Flush()

	return nil
}

func parseCompletionYear(completionDate string) (string, error) {
	completionDate = strings.TrimSpace(completionDate)
	if completionDate == "" {
		return "", fmt.Errorf("empty completion date")
	}

	if len(completionDate) == 4 {
		if _, err := strconv.Atoi(completionDate); err != nil {
			return "", err
		}
		return completionDate, nil
	}

	year, err := time.Parse("2006-01", completionDate)
	if err != nil {
		return "", err
	}

	return year.Format("2006"), nil
}

func getContributors(submission proquest.DISSSubmission) string {
	proquestAuthor := submission.Authorship.Authors[0]
	name := "relators:cre:person:"
	name = name + proquestAuthor.Name.Surname
	name = name + ", " + proquestAuthor.Name.First

	author := islandora.Contributor{
		Name:        name,
		Orcid:       proquestAuthor.ORCiD,
		Institution: submission.Description.Institution.Name,
		Status:      "Graduate Student",
	}

	for _, c := range proquestAuthor.Contacts {
		if c.Email != "" {
			author.Email = c.Email
			break
		}
	}
	var contributors []string
	authorJson, err := json.Marshal(author)
	if err != nil {
		slog.Error("Unable to marshal author", "author", author, "err", err)
		return ""
	}
	contributors = append(contributors, string(authorJson))

	for _, proquestAdvisor := range submission.Description.Advisors {
		name := "relators:ths:person:"
		name = name + proquestAdvisor.Name.Surname
		name = name + ", " + proquestAdvisor.Name.First
		advisor := islandora.Contributor{
			Name:        name,
			Institution: submission.Description.Institution.Name,
			Status:      "Faculty",
		}
		aj, err := json.Marshal(advisor)
		if err != nil {
			slog.Error("Unable to marshal advisor", "advisor", advisor, "err", err)
			return ""
		}

		contributors = append(contributors, string(aj))
	}

	return strings.Join(contributors, " ; ")
}

func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
