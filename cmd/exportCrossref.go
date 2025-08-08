package cmd

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/lehigh-university-libraries/go-islandora/model/crossref"
	"github.com/lehigh-university-libraries/go-islandora/pkg/islandora"
	"github.com/spf13/cobra"
)

var crossrefType,
	crossrefRegistrant,
	crossrefDepositorName,
	crossrefDepositorEmail,
	journalDoi,
	journalUrl,
	journalTitle string

// exportCrossref represents the transformCsvCrossref command
var exportCrossref = &cobra.Command{
	Use:   "crossref",
	Short: "Transform CSV to Crossref",
	Run: func(cmd *cobra.Command, args []string) {
		if target == "" {
			slog.Error("target is required")
			os.Exit(1)
		}
		if baseUrl == "" || nid == 0 {
			slog.Error("--baseUrl and --nid flags are required")
			os.Exit(1)
		}
		if crossrefType != "journal-volume" && crossrefType != "journal-issue" {
			slog.Error("Unsupported type", "type", crossrefType)
			os.Exit(1)
		}

		nodes, err := islandora.FetchNodes(baseUrl, nid)
		if err != nil {
			log.Fatal(err)
		}

		var (
			volumes []crossref.JournalVolume
		)
		for _, node := range nodes {
			if journalDoi == "" || journalUrl == "" {
				for _, id := range *node.FieldIdentifier {
					if id.Attr0 != "doi" || id.Value == "" {
						continue
					}
					journalDoi = id.Value
					journalUrl = fmt.Sprintf("%s/node/%d", baseUrl, nid)
					break
				}
				if journalDoi == "" || journalUrl == "" {
					slog.Error("Unable to find journal DOI or URL")
					os.Exit(1)
				}
				journalTitle = node.Title.String()
				slog.Info("Got journal DOI and url", "doi", journalDoi, "url", journalUrl, "title", journalTitle)
				continue
			}
			journalDoiData := crossref.DoiData{
				Doi: journalDoi,
				Url: journalUrl,
			}
			volume := crossref.JournalVolume{
				JournalTitle:   journalTitle,
				JournalDoiData: journalDoiData,
			}

			currentNid, err := node.Nid.MarshalCSV()
			if err != nil {
				slog.Error("Unable to get nid", "node", node)
				os.Exit(1)
			}
			for _, id := range *node.FieldIdentifier {
				if id.Attr0 != "doi" || id.Value == "" {
					continue
				}
				volume.VolumeDoiData = crossref.DoiData{
					Doi: id.Value,
					Url: fmt.Sprintf("%s/node/%d", baseUrl, currentNid),
				}
			}
			currentNidInt, err := strconv.Atoi(currentNid)
			if err != nil {
				slog.Error("Unable to convert nid to int", "err", err, "nid", currentNid)
				os.Exit(1)
			}

			for _, child := range nodes {
				isArticle := false
				for _, parent := range *child.FieldMemberOf {
					if parent.TargetId == currentNidInt {
						isArticle = true
					}
				}
				if !isArticle {
					continue
				}
				first := true
				var year int
				for _, d := range *child.FieldEdtfDateIssued {
					yearStr := strings.Split(d.Value, "-")[0]
					year, err = strconv.Atoi(yearStr)
					if err != nil {
						slog.Error("Unable to convert year to int", "err", err)
						os.Exit(1)
					}
					volume.Year = year
				}
				article := crossref.Article{
					Title: html.EscapeString(child.FieldFullTitle.String()),
					Year:  year,
				}
				rights := (*child.FieldRights)
				if len(rights) > 0 && !strings.Contains(rights[0].String(), ".getty") {
					article.LicenseRef = rights[0].String()
				}
				abstract := child.FieldAbstract.String()
				if abstract != "" {
					article.Abstract, err = crossref.StrToJATS(abstract)
					if err != nil {
						slog.Error("Unable to convert abstract to JATS", "err", err)
						os.Exit(1)
					}
				}
				for _, agent := range *child.FieldLinkedAgent {
					if agent.RelType == "relators:cre" || agent.RelType == "relators:aut" {
						article.Contributors = append(article.Contributors, crossref.GetContributor(agent.String(), first))
						if first {
							first = false
						}
					}
				}
				for _, id := range *child.FieldIdentifier {
					if id.Attr0 != "doi" {
						continue
					}
					article.DoiData = crossref.DoiData{
						Doi: id.Value,
						Url: fmt.Sprintf("%s/node/%d", baseUrl, currentNid),
					}
					break
				}

				for _, detail := range *node.FieldPartDetail {
					if detail.Type == "volume" {
						volume.Number = detail.Number
					}
					if detail.Type == "issue" {
						volume.Issue = detail.Number
					}
				}
				volume.Articles = append(volume.Articles, article)
			}
			volumes = append(volumes, volume)
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
	exportCmd.AddCommand(exportCrossref)

	exportCrossref.Flags().IntVar(&nid, "nid", 0, "The node ID to export a CSV")
	exportCrossref.Flags().StringVar(&crossrefType, "type", "journal-issue", "Crossref type (book, journal-issue, journal-volume, etc.)")
	exportCrossref.Flags().StringVar(&crossrefRegistrant, "registrant", "", "registrant")
	exportCrossref.Flags().StringVar(&crossrefDepositorName, "depositor-name", "", "Depositor name")
	exportCrossref.Flags().StringVar(&crossrefDepositorEmail, "depositor-email", "", "Depositor email")
	exportCrossref.Flags().StringVar(&journalDoi, "journal-doi", "", "Journal's DOI")
	exportCrossref.Flags().StringVar(&journalUrl, "journal-url", "", "Journal's URL")
	exportCrossref.Flags().StringVar(&target, "target", "", "Where to save target file")

}
