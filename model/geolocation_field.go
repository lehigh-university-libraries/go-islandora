package model

type GeoLocationField struct {
	Latitude     float32 `json:"lat"`
	Longitude    float32 `json:"lng"`
	LatitudeSin  float32 `json:"lat_sin"`
	LatitudeCos  float32 `json:"lat_cos"`
	LongitudeRad float32 `json:"lng_rad"`
	Data         string  `json:"data"`
}
