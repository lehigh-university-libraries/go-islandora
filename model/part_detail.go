package model

import (
	"encoding/json"
	"fmt"
)

type PartDetailField struct {
	Type    string `json:"type,omitempty"`
	Caption string `json:"caption,omitempty"`
	Number  string `json:"number,omitempty"`
	Title   string `json:"title,omitempty"`
}

func (field *PartDetailField) MarshalCSV() (string, error) {
	data, err := json.Marshal(field)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (field *PartDetailField) UnmarshalCSV(csv string) error {
	return json.Unmarshal([]byte(csv), field)
}

func (field *PartDetailField) String() string {
	return fmt.Sprintf("%+v", *field)
}
