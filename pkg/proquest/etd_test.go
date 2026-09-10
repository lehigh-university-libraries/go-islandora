package proquest

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

func TestEmbargoDateXML(t *testing.T) {
	tests := []struct {
		name       string
		code       int
		remove     string
		accept     string
		repository string
		access     string
		local      bool
		want       string
		wantErr    string
	}{
		{name: "no embargo", code: 0, accept: "05/01/2021"},
		{name: "six months", code: 1, accept: "05/01/2021", want: "2021-11-01"},
		{name: "one year", code: 2, accept: "05/01/2021", want: "2022-05-01"},
		{name: "two years", code: 3, accept: "05/01/2021", want: "2023-05-01"},
		{name: "month overflow", code: 1, accept: "08/31/2026", want: "2027-03-03"},
		{name: "leap year overflow", code: 2, accept: "02/29/2024", want: "2025-03-01"},
		{name: "two years from leap day", code: 3, accept: "02/29/2024", want: "2026-03-01"},
		{name: "custom release", code: 4, remove: "08/26/2027", want: "2027-08-26"},
		{name: "release overrides code 0", code: 0, remove: "08/26/2027", want: "2027-08-26"},
		{name: "release overrides code 1", code: 1, remove: "08/26/2027", accept: "05/01/2021", want: "2027-08-26"},
		{name: "release overrides code 2", code: 2, remove: "08/26/2027", accept: "05/01/2021", want: "2027-08-26"},
		{name: "release overrides code 3", code: 3, remove: "08/26/2027", accept: "05/01/2021", want: "2027-08-26"},
		{name: "repository date overrides sales release", code: 4, remove: "08/26/2027", repository: "2028-01-01", want: "2028-01-01"},
		{name: "later sales release overrides repository date", code: 4, remove: "08/26/2027", repository: "2027-01-01", want: "2027-08-26"},
		{name: "later repository timestamp from example", code: 4, remove: "05/20/2028", repository: "2028-07-06 00:00:00", want: "2028-07-06"},
		{name: "later sales release with repository timestamp", code: 4, remove: "08/20/2028", repository: "2028-07-06 00:00:00", want: "2028-08-20"},
		{name: "equal release dates", code: 4, remove: "07/06/2028", repository: "2028-07-06 00:00:00", want: "2028-07-06"},
		{name: "indefinite repository embargo overrides sales release", code: 4, remove: "08/26/2027", repository: "NEVER DELIVER", want: "2999-12-31"},
		{name: "repository date overrides duration", code: 1, accept: "05/01/2021", repository: "2028-01-01", want: "2028-01-01"},
		{name: "later calculated date overrides repository date", code: 3, accept: "05/01/2027", repository: "2028-01-01", want: "2029-05-01"},
		{name: "invalid sales dates preserve repository date", code: 1, remove: "not-a-date", accept: "not-a-date", repository: "2028-01-01", want: "2028-01-01"},
		{name: "repository embargo without sales embargo", code: 0, repository: "2028-01-01", want: "2028-01-01"},
		{name: "repository timestamp", code: 0, repository: "2028-01-01 12:34:56", want: "2028-01-01"},
		{name: "indefinite repository embargo without sales embargo", code: 0, repository: "  Never Deliver  ", want: "2999-12-31"},
		{name: "blank repository falls back to sales release", code: 4, remove: "08/26/2027", repository: "  ", want: "2027-08-26"},
		{name: "invalid repository date cannot fall back", code: 4, remove: "08/26/2027", repository: "not-a-date", wantErr: "DISS_delayed_release"},
		{name: "impossible repository date", code: 0, repository: "2027-02-30", wantErr: "DISS_delayed_release"},
		{name: "repository period needs an explicit date", code: 1, accept: "05/01/2021", repository: "6 months", wantErr: "DISS_delayed_release"},
		{name: "open access preserves sales embargo", code: 4, remove: "08/26/2027", access: "Open access", want: "2027-08-26"},
		{name: "open access preserves indefinite embargo", code: 4, remove: "08/26/2027", repository: "NEVER DELIVER", access: " Open Access ", want: "2999-12-31"},
		{name: "open access preserves later repository date", code: 4, remove: "05/20/2028", repository: "2028-07-06 00:00:00", access: "Open access", want: "2028-07-06"},
		{name: "open access preserves later sales date", code: 4, remove: "08/20/2028", repository: "2028-07-06 00:00:00", access: "Open access", want: "2028-08-20"},
		{name: "open access without embargo", code: 0, access: "Open access"},
		{name: "campus access", code: 0, access: "Campus use only", local: true},
		{name: "campus access preserves sales embargo", code: 4, remove: "08/26/2027", access: "Campus use only", local: true, want: "2027-08-26"},
		{name: "campus access preserves indefinite embargo", code: 0, repository: "NEVER DELIVER", access: " Campus Use Only ", local: true, want: "2999-12-31"},
		{name: "campus access preserves later repository date", code: 4, remove: "05/20/2028", repository: "2028-07-06 00:00:00", access: "Campus use only", local: true, want: "2028-07-06"},
		{name: "campus access preserves later sales date", code: 4, remove: "08/20/2028", repository: "2028-07-06 00:00:00", access: "Campus use only", local: true, want: "2028-08-20"},
		{name: "open access cannot hide invalid repository date", code: 0, repository: "not-a-date", access: "Open access", wantErr: "DISS_delayed_release"},
		{name: "unknown access option", code: 0, access: "Restricted", wantErr: "DISS_access_option"},
		{name: "missing custom release", code: 4, accept: "05/01/2021"},
		{name: "invalid custom release", code: 4, remove: "not-a-date", accept: "05/01/2021"},
		{name: "invalid release falls back to code", code: 1, remove: "not-a-date", accept: "05/01/2021", want: "2021-11-01"},
		{name: "invalid acceptance date", code: 1, accept: "not-a-date"},
		{name: "missing acceptance date", code: 1},
		{name: "negative code", code: -1, wantErr: "invalid embargo code -1"},
		{name: "code above range", code: 5, wantErr: "invalid embargo code 5"},
		{name: "unknown code", code: 99, wantErr: "invalid embargo code 99"},
		{name: "invalid code with release date", code: 5, remove: "08/26/2027", wantErr: "invalid embargo code 5"},
		{name: "invalid code with repository date", code: 5, repository: "2028-01-01", wantErr: "invalid embargo code 5"},
		{name: "invalid code with open access", code: -1, access: "Open access", wantErr: "invalid embargo code -1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf(`<DISS_submission embargo_code="%d">
  <DISS_description><DISS_dates><DISS_accept_date>%s</DISS_accept_date></DISS_dates></DISS_description>
  <DISS_restriction><DISS_sales_restriction code="1" remove="%s"/></DISS_restriction>
  <DISS_repository><DISS_delayed_release>%s</DISS_delayed_release><DISS_access_option>%s</DISS_access_option></DISS_repository>
</DISS_submission>`, tt.code, tt.accept, tt.remove, tt.repository, tt.access)
			var submission DISSSubmission
			if err := xml.Unmarshal([]byte(input), &submission); err != nil {
				t.Fatal(err)
			}
			got, err := submission.EmbargoDate()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("EmbargoDate() error = %v, want %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("EmbargoDate() = %q, want %q", got, tt.want)
			}
			if local := submission.Repository.LocalRestriction(); local != tt.local {
				t.Errorf("LocalRestriction() = %v, want %v", local, tt.local)
			}
		})
	}
}
