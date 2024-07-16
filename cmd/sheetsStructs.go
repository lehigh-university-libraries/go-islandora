package cmd

import (
	"log/slog"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

type CsvColumn struct {
	ColumnName string
	Tag        string
}

// sheetsStructsCmd represents the node-structs command
var sheetsStructsCmd = &cobra.Command{
	Use:   "sheets-structs",
	Short: "Generates Go structs from a Google Sheets template",
	Long: `Generates Go structs from a Google Sheets template,
used to produce Open API specs and related Go code.`,
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")

		fields := sheetsFields()
		structData := StructData{
			StructName: "GoogleSheets",
			CsvColumns: fields,
		}

		structCode, err := generateOapiSpec(structData, "sheets.yaml.tmpl")
		if err != nil {
			slog.Error("Error generating Go struct", "err", err)
			os.Exit(1)
		}

		err = os.WriteFile(output, []byte(structCode), 0644)
		if err != nil {
			slog.Error("Error writing output file", "err", err)
			os.Exit(1)
		}

		slog.Info("Open API Spec generated and written", "file", output)
		genCmd := exec.Command("go", "generate", "./workbench")
		if err := genCmd.Run(); err != nil {
			slog.Error("Unable to run go generate", "err", err)
			os.Exit(1)
		}
		slog.Info("Structs generated in workbench/workbench.gen.go")
	},
}

func init() {
	generateCmd.AddCommand(sheetsStructsCmd)

	sheetsStructsCmd.Flags().String("output", "./workbench.yaml", "Output file for generated Open API spec")
}

func sheetsFields() []CsvColumn {
	fields := []CsvColumn{}
	f := map[string]string{
		"Human Name":                  `"-"`,
		"Upload ID":                   "id",
		"Page/Item Parent ID":         "parent_id",
		"Child Sort Order":            "field_weight",
		"Node ID":                     "nid",
		"Parent Collection":           "field_member_of",
		"Object Model":                "field_model",
		"File Path":                   "file",
		"Add Coverpage (Y/N)":         "field_add_coverpage",
		"Title":                       "title",
		"Full Title":                  "field_full_title",
		"Make Public (Y/N)":           "published",
		"Contributor Name 1":          "field_linked_agent.name",
		"Contributor Relator 1":       "field_linked_agent.rel_type",
		"Contributor Type 1":          "field_linked_agent.vid",
		"ORCID Number 1":              "field_linked_agent.entity.field_identifier.attr0=orcid",
		"Contributor Status 1":        "field_linked_agent.entity.field_contributor_status",
		"Contributor Email 1":         "field_linked_agent.entity.field_email",
		"Contributor Institution 1":   "field_linked_agent.entity.field_relationships",
		"Related Department":          "field_department_name",
		"Resource Type":               "field_resource_type",
		"Genre (Getty AAT)":           "field_genre",
		"Creation Date":               "field_edtf_date_issued",
		"Season":                      "field_date_season",
		"Date Captured":               "field_edtf_date_captured",
		"Embargo Until Date":          "field_edtf_date_embargo",
		"Publisher":                   "field_publisher",
		"Edition":                     "field_edition",
		"Language":                    "field_language",
		"Physical Format (Getty AAT)": "field_physical_form",
		"File Format (MIME Type)":     "field_media_type",
		"Page Count":                  "field_extent.attr0=page",
		"Dimensions":                  "field_extent.attr0=dimensions",
		"File Size":                   "field_extent.attr0=bytes",
		"Run Time (HH:MM:SS)":         "field_extent.attr0=minutes",
		"Digital Origin":              "field_digital_origin",
		"Description":                 "field_abstract.attr0=description",
		"Abstract":                    "field_abstract.attr0=abstract",
		"Preferred-Citation (included only in Fritz Lab and Environmental reports)": "field_note.attr0=preferred-citation",
		"Capture Device":                      "field_note.attr0=capture-device",
		"PPI":                                 "field_note.attr0=ppi",
		"Archival Collection":                 "field_note.attr0=collection",
		"Archival Box":                        "field_note.attr0=box",
		"Archival Series":                     "field_note.attr0=series",
		"Archival Folder":                     "field_note.attr0=folder",
		"Local Restriction":                   "field_local_restriction",
		"Subject Topic (LCSH)":                "field_subject_lcsh",
		"Keyword":                             "field_keywords",
		"Subject Name (LCNAF)":                "field_subjects_name",
		"Subject Geographic (LCNAF)":          "field_geographic_subject.vid=geographic_naf",
		"Subject Geographic (Local)":          "field_geographic_subject.vid=geographic_local",
		"Hierarchical Geographic (Getty TGN)": "field_subject_hierarchical_geo",
		"Source Publication Title":            "field_related_item.title",
		"Source Publication L-ISSN":           "field_related_item.identifier_type=issn",
		"Volume Number":                       "field_part_detail.attr0=volume",
		"Issue Number":                        "field_part_detail.attr0=issue",
		"Page Numbers":                        "field_part_detail.attr0=page",
		"DOI":                                 "field_identifier.attr0=doi",
		"Catalog or ArchivesSpace URL":        "field_identifier.attr0=uri",
		"Call Number":                         "field_identifier.attr0=call-number",
		"Report Number (included only on ATLSS and Fritz Lab spreadsheet)": "field_identifier.attr0=report-number",
		"Rights Statement": "field_rights",
		"Access":           "field_access",
		"LinkedAgent":      "field_linked_agent",
		"Identifier":       "field_identifier",
		"Url":              "url",
	}
	for column, field := range f {
		fields = append(fields, CsvColumn{
			ColumnName: column,
			Tag:        field,
		})
	}

	return fields
}
