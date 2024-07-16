package crossref

type CrossrefHead struct {
	Registrant string
	Depositor  Depositor
	Timestamp  string
	BatchId    string
}

type Depositor struct {
	Name  string
	Email string
}
