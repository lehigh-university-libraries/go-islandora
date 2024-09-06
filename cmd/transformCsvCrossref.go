package cmd

import (
	"bytes"
	"encoding/json"
	"html"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
	"github.com/lehigh-university-libraries/go-islandora/model"
	"github.com/lehigh-university-libraries/go-islandora/model/crossref"
	"github.com/lehigh-university-libraries/go-islandora/workbench"
	"github.com/spf13/cobra"
)

var crossrefType,
	crossrefRegistrant,
	crossrefDepositorName,
	crossrefDepositorEmail,
	journalDoi,
	journalUrl string

// transformCsvCrossrefCmd represents the transformCsvCrossref command
var transformCsvCrossrefCmd = &cobra.Command{
	Use:   "crossref",
	Short: "Transform CSV to Crossref",
	Run: func(cmd *cobra.Command, args []string) {
		if source == "" || target == "" {
			slog.Error("Source and target are required")
			os.Exit(1)
		}

		if crossrefType != "journal-volume" && crossrefType != "journal-issue" {
			slog.Error("Unsupported type", "type", crossrefType)
			os.Exit(1)
		}

		file, err := os.Open(source)
		if err != nil {
			log.Fatalf("could not open the CSV file: %v", err)
		}
		defer file.Close()

		var rows []*workbench.SheetsCsv

		if err := gocsv.UnmarshalFile(file, &rows); err != nil {
			log.Fatalf("could not unmarshal CSV file: %v", err)
		}

		nodeIDMap := make(map[string]*workbench.SheetsCsv)
		for _, row := range rows {
			nodeIDMap[*row.NodeID] = row
		}
		var volumes []crossref.JournalVolume

		journalDoiData := crossref.DoiData{}
		if journalDoi != "" && journalUrl != "" {
			journalDoiData.Doi = journalDoi
			journalDoiData.Url = journalUrl
		} else {
			for _, id := range strings.Split(*rows[0].Identifier, "|") {
				if id == "" {
					continue
				}
				var identifier model.TypedText
				err = json.Unmarshal([]byte(id), &identifier)
				if err != nil {
					slog.Error("Unable to unmarshal journal identifier", "err", err)
					os.Exit(1)
				}
				if identifier.Attr0 == "doi" {
					journalDoiData.Doi = identifier.Value
					journalDoiData.Url = *rows[0].Url
					break
				}
			}
		}
		for _, row := range rows {
			if *row.ObjectModel == "Sub-Collection" {
				volume := crossref.JournalVolume{
					JournalTitle:   *rows[0].Title,
					JournalDoiData: journalDoiData,
				}
				for _, id := range strings.Split(*row.Identifier, "|") {
					if id == "" {
						continue
					}
					var identifier model.TypedText
					err = json.Unmarshal([]byte(id), &identifier)
					if err != nil {
						slog.Error("Unable to unmarshal identifier", "err", err)
						os.Exit(1)
					}
					if identifier.Attr0 == "doi" {
						volume.VolumeDoiData = crossref.DoiData{
							Doi: identifier.Value,
							Url: *row.Url,
						}
					}
				}

				for _, checkRow := range rows {
					if *checkRow.ParentCollection == *row.NodeID {
						first := true
						yearStr := strings.Split(*checkRow.CreationDate, "-")[0]
						year, err := strconv.Atoi(yearStr)
						if err != nil {
							slog.Error("Unable to convert year to int", "err", err)
							os.Exit(1)
						}
						volume.Year = year
						article := crossref.Article{
							Title: html.EscapeString(*checkRow.FullTitle),
							Year:  year,
						}
						if *checkRow.RightsStatement != "" && !strings.Contains(*checkRow.RightsStatement, ".getty") {
							article.LicenseRef = *checkRow.RightsStatement
						}
						if checkRow.FieldAbstract != nil && *checkRow.FieldAbstract != "" {
							var abstract model.TypedText
							err = json.Unmarshal([]byte(*checkRow.FieldAbstract), &abstract)
							if err != nil {
								slog.Error("Unable to unmarshal abstract", "err", err)
								os.Exit(1)
							}
							article.Abstract, err = crossref.StrToJATS(abstract.Value)
							if err != nil {
								slog.Error("Unable to convert abstract to JATS", "err", err)
								os.Exit(1)
							}
						}
						if checkRow.LinkedAgent != nil {
							for _, agent := range strings.Split(*checkRow.LinkedAgent, "|") {
								components := strings.Split(agent, ":")
								if len(components) < 3 {
									continue
								}
								if components[2] != "person" {
									continue
								}
								if components[1] == "cre" || components[1] == "aut" {
									name := strings.Join(components[3:], ":")
									article.Contributors = append(article.Contributors, crossref.GetContributor(name, first))
									if first {
										first = false
									}
								}
							}
						}
						for _, id := range strings.Split(*checkRow.Identifier, "|") {
							if id == "" {
								continue
							}
							var identifier model.TypedText
							err = json.Unmarshal([]byte(id), &identifier)
							if err != nil {
								slog.Error("Unable to unmarshal identifier", "err", err)
								os.Exit(1)
							}
							if identifier.Attr0 == "doi" {
								article.DoiData = crossref.DoiData{
									Doi: identifier.Value,
									Url: *checkRow.Url,
								}
								break
							}
						}

						if volume.Number == "" {
							for _, detail := range strings.Split(*checkRow.PartDetail, "|") {
								var pt model.PartDetail
								err = json.Unmarshal([]byte(detail), &pt)
								if err != nil {
									slog.Error("Unable to unmarshal part detail", "err", err)
									os.Exit(1)
								}
								if pt.Type == "volume" {
									volume.Number = pt.Number
								}
								if pt.Type == "issue" {
									volume.Issue = pt.Number
									break
								}
							}
						}
						if checkRow.References != nil {
							for _, doi := range strings.Split(*checkRow.References, "|") {
								if doi == "" {
									continue
								}
								/* TODO
								reference := crossref.Reference{
									crossref.DoiData{
										Doi: doi,
									},
								}
								article.References = append(article.References, reference)
								*/
							}
						}
						volume.Articles = append(volume.Articles, article)
					}
				}
				volumes = append(volumes, volume)
			}
		}

		journalData := crossref.Journal{
			Head: crossref.CrossrefHead{
				Registrant: crossrefRegistrant,
				Depositor: crossref.Depositor{
					Name:  crossrefDepositorName,
					Email: crossrefDepositorEmail,
				},
				Timestamp: time.Now().Unix(),
				BatchId:   uuid.New().String(),
			},
			JournalVolume: volumes,
		}
		tmplFile := "crossref/journal-volume.xml.tmpl"
		tmpl, err := template.ParseFiles(tmplFile)
		if err != nil {
			slog.Error("Unable to parse template", "err", err)
			os.Exit(1)
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, journalData)
		if err != nil {
			slog.Error("Unable to generate template", "err", err)
			os.Exit(1)
		}

		err = os.WriteFile(target, buf.Bytes(), 0644)
		if err != nil {
			slog.Error("Error writing output file", "err", err)
			os.Exit(1)
		}

		slog.Info("Crossref journal written", "file", target)
	},
}

func init() {
	transformCsvCmd.AddCommand(transformCsvCrossrefCmd)

	transformCsvCrossrefCmd.Flags().StringVar(&crossrefType, "type", "journal-issue", "Crossref type (book, journal-issue, journal-volume, etc.)")
	transformCsvCrossrefCmd.Flags().StringVar(&crossrefRegistrant, "registrant", "", "registrant")
	transformCsvCrossrefCmd.Flags().StringVar(&crossrefDepositorName, "depositor-name", "", "Depositor name")
	transformCsvCrossrefCmd.Flags().StringVar(&crossrefDepositorEmail, "depositor-email", "", "Depositor email")
	transformCsvCrossrefCmd.Flags().StringVar(&journalDoi, "journal-doi", "", "Journal's DOI")
	transformCsvCrossrefCmd.Flags().StringVar(&journalUrl, "journal-url", "", "Journal's URL")

}
