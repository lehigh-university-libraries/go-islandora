package crossref

type Journal struct {
	IssuelessVolumes []IssuelessVolume
	Head             CrossrefHead
}

type IssuelessVolume struct {
	JournalTitle   string
	Type           string
	Number         string
	Year           int
	VolumeDoiData  DoiData `xml:"doi_data"`
	Articles       []Article
	JournalDoiData DoiData `xml:"doi_data"`
}

type DoiData struct {
	Doi string `xml:"doi"`
	Url string `xml:"resource"`
}

type Article struct {
	Title        string
	Abstract     string
	Contributors []Contributor
	DoiData      DoiData `xml:"doi_data"`
	Year         int
	References   []Reference
	LicenseRef   string
}

type Reference struct {
	DoiData DoiData
}

type Contributor struct {
	Name         PersonName
	Suffix       string
	Organization string
	Orcid        string
	Role         string
	Sequence     string
}

type PersonName struct {
	Surname string `xml:"surname"`
	Given   string `xml:"given_name,omitempty"`
}
