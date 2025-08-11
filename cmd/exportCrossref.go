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
	"github.com/lehigh-university-libraries/go-islandora/api"
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
			volumes     []crossref.JournalVolume
			journalNode *api.IslandoraObject
		)

		// First pass: find the journal node and extract journal metadata
		for _, node := range nodes {
			if node.Nid.String() == strconv.Itoa(nid) {
				journalNode = node
				break
			}
		}

		if journalNode == nil {
			slog.Error("Journal node not found", "nid", nid)
			os.Exit(1)
		}

		// Extract journal metadata
		if journalDoi == "" || journalUrl == "" {
			for _, id := range *journalNode.FieldIdentifier {
				if id.Attr0 == "doi" && id.Value != "" {
					journalDoi = id.Value
					journalUrl = fmt.Sprintf("%s/node/%d", baseUrl, nid)
					break
				}
			}
			if journalDoi == "" || journalUrl == "" {
				slog.Error("Unable to find journal DOI or URL")
				os.Exit(1)
			}
			journalTitle = journalNode.Title.String()
			slog.Info("Got journal DOI and url", "doi", journalDoi, "url", journalUrl, "title", journalTitle)
		}

		journalDoiData := crossref.DoiData{
			Doi: journalDoi,
			Url: journalUrl,
		}

		for _, node := range nodes {
			currentNid, err := node.Nid.MarshalCSV()
			if err != nil {
				continue
			}
			currentNidInt, err := strconv.Atoi(currentNid)
			if err != nil {
				continue
			}

			// Check if this node is a direct child of the journal
			isVolumeNode := false
			if node.FieldMemberOf != nil {
				for _, parent := range *node.FieldMemberOf {
					if parent.TargetId == nid {
						isVolumeNode = true
						break
					}
				}
			}

			if !isVolumeNode {
				continue
			}

			volume := crossref.JournalVolume{
				JournalTitle:   journalTitle,
				JournalDoiData: journalDoiData,
			}

			// Extract volume DOI
			for _, id := range *node.FieldIdentifier {
				if id.Attr0 == "doi" && id.Value != "" {
					volume.VolumeDoiData = crossref.DoiData{
						Doi: id.Value,
						Url: fmt.Sprintf("%s/node/%s", baseUrl, currentNid),
					}
					break
				}
			}

			// Extract volume metadata (number, issue)
			if node.FieldPartDetail != nil {
				for _, detail := range *node.FieldPartDetail {
					if detail.Type == "volume" {
						volume.Number = detail.Number
					}
					if detail.Type == "issue" {
						volume.Issue = detail.Number
					}
				}
			}

			// Third pass: find articles that belong to this volume
			var volumeYear int
			for _, childNode := range nodes {
				childNidStr, err := childNode.Nid.MarshalCSV()
				if err != nil {
					continue
				}
				_, err = strconv.Atoi(childNidStr)
				if err != nil {
					continue
				}

				// Check if this child node belongs to the current volume
				isArticleInVolume := false
				if childNode.FieldMemberOf != nil {
					for _, parent := range *childNode.FieldMemberOf {
						if parent.TargetId == currentNidInt {
							isArticleInVolume = true
							break
						}
					}
				}

				if !isArticleInVolume {
					continue
				}

				// Extract article metadata
				article := crossref.Article{
					Title: html.EscapeString(childNode.FieldFullTitle.String()),
				}

				// Extract year from article's date
				if childNode.FieldEdtfDateIssued != nil {
					for _, d := range *childNode.FieldEdtfDateIssued {
						yearStr := strings.Split(d.Value, "-")[0]
						year, err := strconv.Atoi(yearStr)
						if err == nil {
							article.Year = year
							volumeYear = year
						}
						break
					}
				}

				// Extract rights/license
				if childNode.FieldRights != nil {
					rights := (*childNode.FieldRights)
					if len(rights) > 0 && !strings.Contains(rights[0].String(), ".getty") {
						article.LicenseRef = rights[0].String()
					}
				}

				// Extract abstract
				abstract := childNode.FieldAbstract.String()
				if abstract != "" {
					article.Abstract, err = crossref.StrToJATS(abstract)
					if err != nil {
						slog.Error("Unable to convert abstract to JATS", "err", err)
						os.Exit(1)
					}
				}

				// Extract contributors
				first := true
				if childNode.FieldLinkedAgent != nil {
					for _, agent := range *childNode.FieldLinkedAgent {
						if agent.RelType == "relators:cre" || agent.RelType == "relators:aut" {
							article.Contributors = append(article.Contributors, crossref.GetContributor(agent.String(), first))
							if first {
								first = false
							}
						}
					}
				}

				// Extract article DOI and URL
				if childNode.FieldIdentifier != nil {
					for _, id := range *childNode.FieldIdentifier {
						if id.Attr0 == "doi" && id.Value != "" {
							article.DoiData = crossref.DoiData{
								Doi: id.Value,
								Url: fmt.Sprintf("%s/node/%s", baseUrl, childNidStr), // Use child's nid, not volume's
							}
							break
						}
					}
				}

				volume.Articles = append(volume.Articles, article)
			}

			// Set the volume year
			volume.Year = volumeYear

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
