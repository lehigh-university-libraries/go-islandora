package model

type BoolField struct {
	Value bool `json:"value"`
}

type ConfigReferenceField struct {
	TargetId   string `json:"target_id"`
	TargetType string `json:"target_type"`
	TargetUuid string `json:"target_uuid"`
}

type EdtfField struct {
	Value string `json:"value"`
}

type EmailField struct {
	Value string `json:"value"`
}

type EntityReferenceField struct {
	TargetId   int    `json:"target_id"`
	TargetType string `json:"target_type"`
	TargetUuid string `json:"target_uuid"`
	Url        string `json:"url"`
}

type GenericField struct {
	Format string `json:"format,omitempty"`
	Value  string `json:"value"`
}

type GeoLocationField struct {
	Latitude     float32 `json:"lat"`
	Longitude    float32 `json:"lng"`
	LatitudeSin  float32 `json:"lat_sin"`
	LatitudeCos  float32 `json:"lat_cos"`
	LongitudeRad float32 `json:"lng_rad"`
	Data         string  `json:"data"`
}

type HierarchicalGeographicField struct {
	City      string `json:"city,omitempty"`
	Continent string `json:"continent,omitempty"`
	Country   string `json:"country,omitempty"`
	County    string `json:"county,omitempty"`
	State     string `json:"state,omitempty"`
	Territory string `json:"territory,omitempty"`
}

type IntField struct {
	Value int `json:"value"`
}

type PartDetailField struct {
	Type    string `json:"type,omitempty"`
	Caption string `json:"caption,omitempty"`
	Number  string `json:"number,omitempty"`
	Title   string `json:"title,omitempty"`
}

type RelatedItemField struct {
	Identifier string `json:"identifier,omitempty"`
	Title      string `json:"title,omitempty"`
	Number     string `json:"number,omitempty"`
}

type TypedRelationField struct {
	TargetId int    `json:"target_id"`
	RelType  string `json:"rel_type"`
}

type TypedTextField struct {
	Attr0  string `json:"attr0,omitempty"`
	Attr1  string `json:"attr1,omitempty"`
	Format string `json:"format,omitempty"`
	Value  string `json:"value"`
}
