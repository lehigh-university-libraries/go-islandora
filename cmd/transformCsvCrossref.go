package cmd

import (
	"bytes"
	"encoding/json"
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
	crossrefDepositorEmail string

// transformCsvCrossrefCmd represents the transformCsvCrossref command
var transformCsvCrossrefCmd = &cobra.Command{
	Use:   "crossref",
	Short: "Transform CSV to Crossref",
	Run: func(cmd *cobra.Command, args []string) {
		if source == "" || target == "" {
			slog.Error("Source and target are required")
			os.Exit(1)
		}

		if crossrefType != "issueless-journal" {
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
		for _, row := range rows[:1] {
			nodeIDMap[*row.NodeID] = row
		}
		var volumes []crossref.IssuelessVolume
		for _, row := range rows[2:] {
			if *row.ObjectModel == "Sub-Collection" {
				volume := crossref.IssuelessVolume{
					JournalTitle: *rows[1].Title,
				}
				for _, id := range strings.Split(*row.Identifier, "|") {
					if id == "" {
						continue
					}
					var identifier model.TypedText
					slog.Info("ID", "id", id)
					err = json.Unmarshal([]byte(id), &identifier)
					if err != nil {
						slog.Error("Unable to unmarshal identifier", "err", err)
						os.Exit(1)
					}
					if identifier.Attr0 == "doi" {
						volume.DoiData = crossref.DoiData{
							Doi: identifier.Value,
							Url: *row.Url,
						}
					}
				}

				for _, checkRow := range rows {
					if *checkRow.ParentCollection == *row.NodeID {
						first := true
						year, err := strconv.Atoi(*checkRow.CreationDate)
						if err != nil {
							slog.Error("Unable to convert year to int", "err", err)
							os.Exit(1)
						}
						volume.Year = year
						article := crossref.Article{
							Title: *checkRow.FullTitle,
							Year:  year,
						}
						if *checkRow.RightsStatement != "" && !strings.Contains(*checkRow.RightsStatement, ".getty") {
							article.LicenseRef = *checkRow.RightsStatement
						}
						if *checkRow.FieldAbstract != "" {
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
								nameComponents := strings.Split(name, ", ")
								surname := nameComponents[0]
								given := ""
								if len(nameComponents) > 1 {
									given = strings.Join(nameComponents[1:], ", ")
								}
								sequence := "additional"
								if first {
									first = false
									sequence = "first"
								}
								article.Contributors = append(article.Contributors, crossref.Contributor{
									Name: crossref.PersonName{
										Given:   given,
										Surname: surname,
									},
									Role:     "author",
									Sequence: sequence,
								})
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

						for _, doi := range strings.Split(*checkRow.References, "|") {
							if doi == "" {
								continue
							}
							reference := crossref.Reference{
								crossref.DoiData{
									Doi: doi,
								},
							}
							article.References = append(article.References, reference)
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
			DoiData: crossref.DoiData{
				Doi: *rows[1].Url,
				Url: *rows[1].Url,
			},
			IssuelessVolumes: volumes,
		}
		tmpl, err := template.ParseFiles("crossref/issueless-journal.xml.tmpl")
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

	transformCsvCrossrefCmd.Flags().StringVar(&crossrefType, "type", "issueless-journal", "Crossref type (book, journal, etc.)")
	transformCsvCrossrefCmd.Flags().StringVar(&crossrefRegistrant, "registrant", "", "registrant")
	transformCsvCrossrefCmd.Flags().StringVar(&crossrefDepositorName, "depositor-name", "", "Depositor name")
	transformCsvCrossrefCmd.Flags().StringVar(&crossrefDepositorEmail, "depositor-email", "", "Depositor email")
}
