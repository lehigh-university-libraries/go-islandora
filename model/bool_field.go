package model

type BoolField struct {
	Value bool `json:"value"`
}

func (field *BoolField) MarshalCSV() (string, error) {
	return field.String(), nil
}

func (field *BoolField) UnmarshalCSV(csv string) error {
	if csv == "1" || csv == "true" {
		field.Value = true
	} else {
		field.Value = false
	}

	return nil
}

func (field *BoolField) String() string {
	if field.Value {
		return "1"
	}

	return "0"
}
