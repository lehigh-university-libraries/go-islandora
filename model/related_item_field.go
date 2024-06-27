package model

import (
	"encoding/json"
	"fmt"
)

type RelatedItemField struct {
	Identifier string `json:"identifier,omitempty"`
	Title      string `json:"title,omitempty"`
	Number     string `json:"number,omitempty"`
}

func (field *RelatedItemField) MarshalCSV() (string, error) {
	data, err := json.Marshal(field)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (field *RelatedItemField) UnmarshalCSV(csv string) error {
	return json.Unmarshal([]byte(csv), field)
}

func (field *RelatedItemField) String() string {
	return fmt.Sprintf("%+v", *field)
}
