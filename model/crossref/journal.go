package crossref

type Journal struct {
	JournalVolume []JournalVolume
	Head          CrossrefHead
}

type JournalVolume struct {
	JournalTitle   string
	Issue          string
	Type           string
	Number         string
	Year           int
	VolumeDoiData  DoiData
	Articles       []Article
	JournalDoiData DoiData
}

type DoiData struct {
	Doi string `xml:"doi"`
	Url string `xml:"resource"`
}

type Article struct {
	Title        string
	Abstract     string
	Contributors []Contributor
	DoiData      DoiData
	Year         int
	References   []Reference
	LicenseRef   string
}

type Reference struct {
	DoiData DoiData
}
