package cmd

import (
	"bytes"
	"log/slog"
	"os"
	"text/template"

	"github.com/lehigh-university-libraries/go-islandora/model/crossref"
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

		journalData := crossref.Journal{
			Head: crossref.CrossrefHead{
				Registrant: crossrefRegistrant,
				Depositor: crossref.Depositor{
					Name:  crossrefDepositorName,
					Email: crossrefDepositorEmail,
				},
			},
			DoiData: crossref.DoiData{
				Doi: "pbe",
				Url: "https://joecorall.com/volume/7",
			},
			IssuelessVolumes: []crossref.IssuelessVolume{
				crossref.IssuelessVolume{
					JournalTitle: "Perspectives on Business and Economics",
					Year:         2009,
					Number:       "12",
					DoiData: crossref.DoiData{
						Doi: "pbe-v123",
						Url: "https://joecorall.com/volume/7",
					},
					Articles: []crossref.Article{
						crossref.Article{
							Contributors: []crossref.Contributor{
								crossref.Contributor{
									Name: crossref.PersonName{
										Surname: "Corall",
										Given:   "Joe",
									},
									Role:     "author",
									Sequence: "first",
								},
							},
							Year: 2009,
							DoiData: crossref.DoiData{
								Doi: "pbe-v123-a456",
								Url: "https://joecorall.com/article/12",
							},
						},
					},
				},
			},
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
