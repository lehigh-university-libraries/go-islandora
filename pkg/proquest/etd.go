package proquest

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type DISSSubmission struct {
	XMLName          xml.Name             `xml:"DISS_submission"`
	EmbargoCode      int                  `xml:"embargo_code,attr"`
	Authorship       DISSEAuthorship      `xml:"DISS_authorship"`
	Description      DISSDescription      `xml:"DISS_description"`
	Repository       DISSRepository       `xml:"DISS_repository"`
	SalesRestriction DISSSalesRestriction `xml:"DISS_restriction>DISS_sales_restriction"`
	Content          DISSContent          `xml:"DISS_content"`
}

// DISSSalesRestriction contains the ProQuest embargo removal date.
type DISSSalesRestriction struct {
	Remove string `xml:"remove,attr"`
}

type DISSRepository struct {
	// These restrictions apply to the university repository, independently of ProQuest sales.
	Embargo      string `xml:"DISS_delayed_release"`
	AccessOption string `xml:"DISS_access_option"`
}

// LocalRestriction reports whether repository access is limited to campus users.
func (repository DISSRepository) LocalRestriction() bool {
	return strings.EqualFold(strings.TrimSpace(repository.AccessOption), "Campus use only")
}

type DISSEAuthorship struct {
	Authors []DISSAuthor `xml:"DISS_author"`
}

type DISSAuthor struct {
	Type        string        `xml:"type,attr"`
	Citizenship string        `xml:"DISS_citizenship,omitempty"`
	Name        DISSName      `xml:"DISS_name"`
	Contacts    []DISSContact `xml:"DISS_contact"`
	ORCiD       string        `xml:"DISS_orcid"`
}

type DISSName struct {
	Surname string `xml:"DISS_surname"`
	First   string `xml:"DISS_fname"`
	Middle  string `xml:"DISS_middle,omitempty"`
	Suffix  string `xml:"DISS_suffix,omitempty"`
}

type DISSContact struct {
	Type    string       `xml:"type,attr"`
	Email   string       `xml:"DISS_email"`
	Address DISSAddress  `xml:"DISS_address"`
	Phone   DISSPhoneFax `xml:"DISS_phone_fax"`
}

type DISSAddress struct {
	Line    string `xml:"DISS_addrline"`
	City    string `xml:"DISS_city"`
	State   string `xml:"DISS_st"`
	Zip     string `xml:"DISS_pcode"`
	Country string `xml:"DISS_country"`
}

type DISSPhoneFax struct {
	Type     string `xml:"type,attr"`
	Country  string `xml:"DISS_cntry_cd"`
	AreaCode string `xml:"DISS_area_code"`
	Number   string `xml:"DISS_phone_num"`
	Ext      string `xml:"DISS_phone_ext,omitempty"`
}

type DISSDescription struct {
	Title          string             `xml:"DISS_title"`
	Degree         string             `xml:"DISS_degree"`
	DegreeLevel    string             `xml:"ETD-Degree-Level"`
	Discipline     string             `xml:"ETD-Degree-Discipline"`
	Institution    DISSInstitution    `xml:"DISS_institution"`
	PageCount      int                `xml:"page_count,attr"`
	Department     string             `xml:"lehigh_departments/lehigh_department"`
	Advisors       []DISSAdvisor      `xml:"DISS_advisor"`
	Categorization DISSCategorization `xml:"DISS_categorization"`
	Dates          DISSDates          `xml:"DISS_dates"`
}

type DISSInstitution struct {
	Name       string `xml:"DISS_inst_name"`
	Department string `xml:"DISS_inst_contact"`
}

type DISSAdvisor struct {
	Name DISSName `xml:"DISS_name"`
}

type DISSCategorization struct {
	Categories []DISSCategory `xml:"DISS_category"`
	Keywords   []string       `xml:"DISS_keyword"`
	Language   string         `xml:"DISS_language"`
}

type DISSCategory struct {
	Description string `xml:"DISS_cat_desc"`
}

type DISSDates struct {
	AcceptDate     string `xml:"DISS_accept_date"`
	CompletionDate string `xml:"DISS_comp_date"`
}

type DISSContent struct {
	Abstract DISSAbstract `xml:"DISS_abstract"`
	Binary   DISSBinary   `xml:"DISS_binary"`
}

type DISSAbstract struct {
	Paragraphs []string `xml:"DISS_para"`
}

type DISSBinary struct {
	Type     string `xml:"type,attr"`
	FileName string `xml:",chardata"`
}

// EmbargoDate returns the later of the repository and ProQuest embargo dates.
// Access options do not clear embargoes; an indefinite restriction takes precedence.
func (submission DISSSubmission) EmbargoDate() (string, error) {
	if submission.EmbargoCode < 0 || submission.EmbargoCode > 4 {
		return "", fmt.Errorf("invalid embargo code %d: expected 0-4", submission.EmbargoCode)
	}

	switch strings.ToLower(strings.TrimSpace(submission.Repository.AccessOption)) {
	case "", "open access", "campus use only":
	default:
		return "", fmt.Errorf("DISS_access_option %q cannot be represented by an embargo date", submission.Repository.AccessOption)
	}

	var repositoryDate string
	if embargo := strings.TrimSpace(submission.Repository.Embargo); embargo != "" {
		if strings.EqualFold(embargo, "never deliver") {
			// Preserve Lehigh's existing convention for an indefinite repository embargo.
			return "2999-12-31", nil
		}
		for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05"} {
			if date, err := time.Parse(layout, embargo); err == nil {
				repositoryDate = date.Format("2006-01-02")
				break
			}
		}
		if repositoryDate == "" {
			return "", fmt.Errorf("cannot determine repository embargo date from DISS_delayed_release %q", embargo)
		}
	}

	releaseDate, err := time.Parse("01/02/2006", submission.SalesRestriction.Remove)
	if err == nil {
		return max(repositoryDate, releaseDate.Format("2006-01-02")), nil
	}

	if submission.EmbargoCode == 0 || submission.EmbargoCode == 4 {
		return repositoryDate, nil
	}

	acceptDate, err := time.Parse("01/02/2006", submission.Description.Dates.AcceptDate)
	if err != nil {
		slog.Error("Invalid acceptance date format", "date", submission.Description.Dates.AcceptDate, "error", err)
		return repositoryDate, nil
	}

	switch submission.EmbargoCode {
	case 1:
		releaseDate = acceptDate.AddDate(0, 6, 0)
	case 2:
		releaseDate = acceptDate.AddDate(1, 0, 0)
	case 3:
		releaseDate = acceptDate.AddDate(2, 0, 0)
	}

	return max(repositoryDate, releaseDate.Format("2006-01-02")), nil
}
