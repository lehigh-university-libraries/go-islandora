package model

import (
	"encoding/json"
	"fmt"
)

type TypedTextField struct {
	Attr0  string `json:"attr0,omitempty"`
	Attr1  string `json:"attr1,omitempty"`
	Format string `json:"format,omitempty"`
	Value  string `json:"value"`
}

func (field *TypedTextField) MarshalCSV() (string, error) {
	data, err := json.Marshal(field)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (field *TypedTextField) UnmarshalCSV(csv string) error {
	return json.Unmarshal([]byte(csv), field)
}

func (field *TypedTextField) String() string {
	return fmt.Sprintf("%+v", *field)
}
