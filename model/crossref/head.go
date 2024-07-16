package crossref

type CrossrefHead struct {
	Registrant string
	Depositor  Depositor
	Timestamp  int64
	BatchId    string
}

type Depositor struct {
	Name  string
	Email string
}
