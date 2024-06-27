package model

import (
	"encoding/json"
	"fmt"
)

type GenericField struct {
	Format    string `json:"format,omitempty"`
	Processed string `json:"processed,omitempty"`
	Value     string `json:"value"`
}

func (field *GenericField) MarshalCSV() (string, error) {
	data, err := json.Marshal(field)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (field *GenericField) UnmarshalCSV(csv string) error {
	return json.Unmarshal([]byte(csv), field)
}

func (field *GenericField) String() string {
	return fmt.Sprintf("%+v", *field)
}
