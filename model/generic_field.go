package model

type GenericField struct {
	Format    string `json:"format,omitempty"`
	Processed string `json:"processed,omitempty"`
	Value     string `json:"value"`
}
