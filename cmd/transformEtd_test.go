package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestETDCommandsRejectInvalidEmbargo(t *testing.T) {
	oldSource, oldTarget, oldCSV := source, target, etdCsvFile
	t.Cleanup(func() { source, target, etdCsvFile = oldSource, oldTarget, oldCSV })
	source = t.TempDir()
	target = filepath.Join(t.TempDir(), "output.csv")
	etdCsvFile = filepath.Join(t.TempDir(), "export.csv")
	if err := os.WriteFile(etdCsvFile, []byte("title,nid,field_edtf_date_issued_value\n"), 0600); err != nil {
		t.Fatal(err)
	}

	var archive bytes.Buffer
	w := zip.NewWriter(&archive)
	f, err := w.Create("test_DATA.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(`<DISS_submission embargo_code="5"/>`)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "test.zip"), archive.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	for _, command := range []*cobra.Command{transformEtdCmd, transformEtdDateBackfillCmd} {
		t.Run(command.Name(), func(t *testing.T) {
			if command.RunE == nil {
				t.Fatal("command must return errors")
			}
			err := command.RunE(command, nil)
			if err == nil || !strings.Contains(err.Error(), "invalid embargo code 5") {
				t.Fatalf("command error = %v, want invalid embargo code 5", err)
			}
		})
	}
}

func TestParseCompletionYear(t *testing.T) {
	tests := []struct {
		name           string
		completionDate string
		want           string
		wantErr        bool
	}{
		{
			name:           "year and month",
			completionDate: "2009-05",
			want:           "2009",
		},
		{
			name:           "year only",
			completionDate: "2010",
			want:           "2010",
		},
		{
			name:           "trims whitespace",
			completionDate: " 2009 ",
			want:           "2009",
		},
		{
			name:           "invalid year only",
			completionDate: "20AB",
			wantErr:        true,
		},
		{
			name:           "invalid month",
			completionDate: "2009-13",
			wantErr:        true,
		},
		{
			name:           "empty",
			completionDate: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCompletionYear(tt.completionDate)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseCompletionYear(%q) expected error", tt.completionDate)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCompletionYear(%q) unexpected error: %v", tt.completionDate, err)
			}
			if got != tt.want {
				t.Fatalf("parseCompletionYear(%q) = %q, want %q", tt.completionDate, got, tt.want)
			}
		})
	}
}

func TestETDCampusRestrictionOutputs(t *testing.T) {
	oldSource, oldTarget, oldCSV := source, target, etdCsvFile
	t.Cleanup(func() { source, target, etdCsvFile = oldSource, oldTarget, oldCSV })
	source = t.TempDir()
	target = filepath.Join(t.TempDir(), "ingest.csv")
	etdCsvFile = filepath.Join(t.TempDir(), "export.csv")
	const input = `<DISS_submission embargo_code="4">
  <DISS_authorship><DISS_author><DISS_name><DISS_surname>Test</DISS_surname></DISS_name></DISS_author></DISS_authorship>
  <DISS_description external_id="http://dissertations.umi.com/lehigh:12345"><DISS_title>Campus thesis</DISS_title><DISS_dates><DISS_comp_date>2026-08</DISS_comp_date></DISS_dates><DISS_categorization><DISS_language>en</DISS_language></DISS_categorization></DISS_description>
  <DISS_restriction><DISS_sales_restriction remove="08/26/2027"/></DISS_restriction>
  <DISS_repository><DISS_access_option>Campus use only</DISS_access_option></DISS_repository>
</DISS_submission>`
	var archive bytes.Buffer
	w := zip.NewWriter(&archive)
	f, err := w.Create("campus_DATA.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "campus.zip"), archive.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := transformETDs(&cobra.Command{}, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil || len(rows) != 2 {
		t.Fatalf("ingest CSV = %v, error %v", rows, err)
	}
	fields := map[string]string{}
	for i, name := range rows[0] {
		fields[name] = rows[1][i]
	}
	if fields["Local Restriction"] != "true" || fields["Embargo Until Date"] != "2027-08-26" {
		t.Fatalf("incorrect campus restrictions: %v", fields)
	}
	if _, exists := fields["Identifier"]; exists {
		t.Fatal("ingest CSV must not include ProQuest identifiers")
	}

	command := &cobra.Command{}
	for _, tt := range []struct {
		title, existing string
		wantUpdate      bool
	}{
		{"Campus thesis", "false", true},
		{"Campus thesis", "true", false},
		{"Changed title", "false", false},
	} {
		var input bytes.Buffer
		writer := csv.NewWriter(&input)
		if err := writer.WriteAll([][]string{
			{"title", "nid", "field_edtf_date_issued_value", "field_edtf_date_embargo_value", "field_local_restriction"},
			{tt.title, "42", "2026", "2027-08-26", tt.existing},
		}); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(etdCsvFile, input.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		command.SetOut(&out)
		if err := etdDateBackfill(command, nil); err != nil {
			t.Fatal(err)
		}
		want := "nid\tfield_edtf_date_issued_value\tfield_edtf_date_embargo_value\tfield_local_restriction\n"
		if tt.wantUpdate {
			want += "42\t2026\t2027-08-26\ttrue\n"
		}
		if out.String() != want {
			t.Errorf("backfill for %+v = %q, want %q", tt, out.String(), want)
		}
	}
}
