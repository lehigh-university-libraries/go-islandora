package model

type EdtfField struct {
	Value string `json:"value"`
}

func (field *EdtfField) MarshalCSV() (string, error) {
	return field.Value, nil
}

func (field *EdtfField) UnmarshalCSV(csv string) error {
	field.Value = csv

	return nil
}

func (field *EdtfField) String() string {
	return field.Value
}
