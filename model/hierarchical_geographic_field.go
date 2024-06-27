package model

type HierarchicalGeographicField struct {
	City      string `json:"city,omitempty"`
	Continent string `json:"continent,omitempty"`
	Country   string `json:"country,omitempty"`
	County    string `json:"county,omitempty"`
	State     string `json:"state,omitempty"`
	Territory string `json:"territory,omitempty"`
}
