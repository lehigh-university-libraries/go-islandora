package model

import (
	"strconv"
)

type IntField struct {
	Value int `json:"value"`
}

func (field *IntField) MarshalCSV() (string, error) {
	return strconv.Itoa(field.Value), nil
}

func (field *IntField) UnmarshalCSV(csv string) error {
	var err error
	field.Value, err = strconv.Atoi(csv)

	return err
}

func (field *IntField) String() string {
	return strconv.Itoa(field.Value)
}
