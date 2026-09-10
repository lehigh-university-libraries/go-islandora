package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestNeverDeliverWithEmptySalesRelease(t *testing.T) {
	oldSource, oldTarget, oldFutureOnly := source, target, etdEmbargoFutureOnly
	t.Cleanup(func() { source, target, etdEmbargoFutureOnly = oldSource, oldTarget, oldFutureOnly })
	source, target = t.TempDir(), ""
	input, err := os.ReadFile("testdata/etd-never-deliver.xml")
	if err != nil {
		t.Fatal(err)
	}
	// Exercise a raw ISO-8859-1 byte, keeping the checked-in fixture ASCII-readable.
	input = bytes.ReplaceAll(input, []byte("&#225;"), []byte{0xe1})
	xmlPath := filepath.Join(source, "Brooks_DATA.xml")
	if err := os.WriteFile(xmlPath, input, 0600); err != nil {
		t.Fatal(err)
	}
	const pdf = "Brooks_lehigh_0105N_12752.pdf"
	for _, futureOnly := range []bool{false, true} {
		etdEmbargoFutureOnly = futureOnly
		var out bytes.Buffer
		command := &cobra.Command{}
		command.SetOut(&out)
		if err := extractETDEmbargoes(command, nil); err != nil {
			t.Fatal(err)
		}
		rows, err := csv.NewReader(&out).ReadAll()
		if err != nil || len(rows) != 2 {
			t.Fatalf("future-only=%v: CSV = %v, error %v", futureOnly, rows, err)
		}
		want := []string{xmlPath, "2999-12-31", "false", pdf}
		if !reflect.DeepEqual(rows[1], want) {
			t.Errorf("future-only=%v: row = %v, want %v", futureOnly, rows[1], want)
		}
	}

	var archive bytes.Buffer
	w := zip.NewWriter(&archive)
	for name, data := range map[string][]byte{"Brooks_DATA.xml": input, pdf: nil} {
		file, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(source, "Brooks.zip")
	if err := os.WriteFile(zipPath, archive.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	target = filepath.Join(t.TempDir(), "ingest.csv")
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
	for field, want := range map[string]string{
		"Embargo Until Date": "2999-12-31",
		"Local Restriction":  "false",
		"Creation Date":      "2023",
		"File Path":          filepath.Join(source, pdf),
		"Abstract":           "<p>Naylor’s manuscript features Sámi Indigenous groups.</p>",
	} {
		if fields[field] != want {
			t.Errorf("%s = %q, want %q", field, fields[field], want)
		}
	}

	var backfill bytes.Buffer
	records := map[string]etdRecord{
		fields["Title"]: {nid: "42", fieldEdtfDateIssuedValue: "2023"},
	}
	if err := processZipForDateCheck(zipPath, records, &backfill); err != nil {
		t.Fatal(err)
	}
	if want := "42\t2023\t2999-12-31\tfalse\n"; backfill.String() != want {
		t.Errorf("backfill = %q, want %q", backfill.String(), want)
	}
}

func TestExtractETDEmbargoes(t *testing.T) {
	oldSource, oldTarget := source, target
	t.Cleanup(func() { source, target = oldSource, oldTarget })
	source, target = t.TempDir(), ""
	if err := os.Mkdir(filepath.Join(source, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	inputs := map[string]string{
		"campus.xml":         `<DISS_submission embargo_code="4"><DISS_restriction><DISS_sales_restriction remove="08/26/2027"/></DISS_restriction><DISS_repository><DISS_access_option>Campus use only</DISS_access_option></DISS_repository></DISS_submission>`,
		"nested/release.XML": `<DISS_submission embargo_code="4"><DISS_description external_id="http://dissertations.umi.com/lehigh:12345"/><DISS_restriction><DISS_sales_restriction remove="08/26/2027"/></DISS_restriction><DISS_repository><DISS_delayed_release>2027-01-01 00:00:00</DISS_delayed_release></DISS_repository><DISS_content><DISS_binary type="PDF">Amin_lehigh_0105A_13188.pdf</DISS_binary></DISS_content></DISS_submission>`,
		"repository.xml":     `<DISS_submission embargo_code="0"><DISS_repository><DISS_delayed_release>NEVER DELIVER</DISS_delayed_release></DISS_repository></DISS_submission>`,
		"ignored.txt":        "not XML",
	}
	for name, input := range inputs {
		if err := os.WriteFile(filepath.Join(source, name), []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	command := &cobra.Command{}
	command.SetOut(&out)
	if err := extractETDEmbargoes(command, nil); err != nil {
		t.Fatal(err)
	}
	got, err := csv.NewReader(&out).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"file", "field_edtf_date_embargo", "field_local_restriction", "pdf_file"},
		{filepath.Join(source, "campus.xml"), "2027-08-26", "true", ""},
		{filepath.Join(source, "nested/release.XML"), "2027-08-26", "false", "Amin_lehigh_0105A_13188.pdf"},
		{filepath.Join(source, "repository.xml"), "2999-12-31", "false", ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CSV = %v, want %v", got, want)
	}

	target = filepath.Join(t.TempDir(), "embargoes.csv")
	if err := extractETDEmbargoes(command, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	got, err = csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("target CSV = %v, error %v, want %v", got, err, want)
	}
}

func TestExtractETDEmbargoesFutureOnly(t *testing.T) {
	oldSource, oldTarget, oldFutureOnly := source, target, etdEmbargoFutureOnly
	t.Cleanup(func() { source, target, etdEmbargoFutureOnly = oldSource, oldTarget, oldFutureOnly })
	source, target = t.TempDir(), ""
	now := time.Now().UTC()
	for name, date := range map[string]string{
		"past":       now.AddDate(0, 0, -1).Format("2006-01-02"),
		"today":      now.Format("2006-01-02"),
		"future":     now.AddDate(0, 0, 1).Format("2006-01-02"),
		"indefinite": "NEVER DELIVER",
		"none":       "",
	} {
		input := fmt.Sprintf(`<DISS_submission embargo_code="0"><DISS_repository><DISS_delayed_release>%s</DISS_delayed_release></DISS_repository></DISS_submission>`, date)
		if err := os.WriteFile(filepath.Join(source, name+".xml"), []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, futureOnly := range []bool{false, true} {
		etdEmbargoFutureOnly = futureOnly
		var out bytes.Buffer
		command := &cobra.Command{}
		command.SetOut(&out)
		if err := extractETDEmbargoes(command, nil); err != nil {
			t.Fatal(err)
		}
		rows, err := csv.NewReader(&out).ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		var files []string
		for _, row := range rows[1:] {
			files = append(files, filepath.Base(row[0]))
		}
		want := []string{"future.xml", "indefinite.xml", "none.xml", "past.xml", "today.xml"}
		if futureOnly {
			want = []string{"future.xml", "indefinite.xml"}
		}
		if !reflect.DeepEqual(files, want) {
			t.Errorf("future-only=%v: files = %v, want %v", futureOnly, files, want)
		}
	}
}

func TestExtractETDEmbargoesErrors(t *testing.T) {
	oldSource, oldTarget, oldFutureOnly := source, target, etdEmbargoFutureOnly
	t.Cleanup(func() { source, target, etdEmbargoFutureOnly = oldSource, oldTarget, oldFutureOnly })
	etdEmbargoFutureOnly = true
	for _, tt := range []struct{ name, input, want string }{
		{"unknown code", `<DISS_submission embargo_code="5"/>`, "invalid embargo code 5"},
		{"invalid XML", `<DISS_submission>`, "failed to decode XML"},
		{"unresolved repository restriction", `<DISS_submission embargo_code="0"><DISS_repository><DISS_delayed_release>6 months</DISS_delayed_release></DISS_repository></DISS_submission>`, "DISS_delayed_release"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source, target = t.TempDir(), ""
			path := filepath.Join(source, "invalid.xml")
			if err := os.WriteFile(path, []byte(tt.input), 0600); err != nil {
				t.Fatal(err)
			}
			command := &cobra.Command{}
			command.SetOut(&bytes.Buffer{})
			err := extractETDEmbargoes(command, nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) || !strings.Contains(err.Error(), path) {
				t.Fatalf("error = %v, want %q and file path", err, tt.want)
			}
		})
	}
}
