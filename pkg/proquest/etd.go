package proquest

import (
	"encoding/xml"
)

type DISSSubmission struct {
	XMLName     xml.Name        `xml:"DISS_submission"`
	EmbargoCode int             `xml:"embargo_code,attr"`
	Authorship  DISSEAuthorship `xml:"DISS_authorship"`
	Description DISSDescription `xml:"DISS_description"`
	Restriction DISSRestriction `xml:"DISS_restriction"`
	Content     DISSContent     `xml:"DISS_content"`
}

type DISSRestriction struct {
	SalesRestriction DISSSalesRestriction `xml:"DISS_sales_restriction"`
}

type DISSSalesRestriction struct {
	Remove string `xml:"remove,attr"`
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
