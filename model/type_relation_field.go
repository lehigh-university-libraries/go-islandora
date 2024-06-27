package model

import (
	"encoding/json"
	"fmt"
)

type TypedRelationField struct {
	TargetId int    `json:"target_id"`
	RelType  string `json:"rel_type"`
}

func (field *TypedRelationField) MarshalCSV() (string, error) {
	data, err := json.Marshal(field)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (field *TypedRelationField) UnmarshalCSV(csv string) error {
	return json.Unmarshal([]byte(csv), field)
}

func (field *TypedRelationField) String() string {
	return fmt.Sprintf("%+v", *field)
}
