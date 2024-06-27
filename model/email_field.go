package model

type EmailField struct {
	Value string `json:"value"`
}

func (field *EmailField) MarshalCSV() (string, error) {
	return field.Value, nil
}

func (field *EmailField) UnmarshalCSV(csv string) error {
	field.Value = csv

	return nil
}

func (field *EmailField) String() string {
	return field.Value
}
