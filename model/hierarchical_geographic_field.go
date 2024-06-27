package model

import (
	"encoding/json"
	"fmt"
)

type HierarchicalGeographicField struct {
	City      string `json:"city,omitempty"`
	Continent string `json:"continent,omitempty"`
	Country   string `json:"country,omitempty"`
	County    string `json:"county,omitempty"`
	State     string `json:"state,omitempty"`
	Territory string `json:"territory,omitempty"`
}

func (field *HierarchicalGeographicField) MarshalCSV() (string, error) {
	data, err := json.Marshal(field)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (field *HierarchicalGeographicField) UnmarshalCSV(csv string) error {
	return json.Unmarshal([]byte(csv), field)
}

func (field *HierarchicalGeographicField) String() string {
	return fmt.Sprintf("%+v", *field)
}
