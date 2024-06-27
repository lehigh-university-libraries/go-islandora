package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type GeoLocationField struct {
	Latitude     float32 `json:"lat"`
	Longitude    float32 `json:"lng"`
	LatitudeSin  float32 `json:"lat_sin"`
	LatitudeCos  float32 `json:"lat_cos"`
	LongitudeRad float32 `json:"lng_rad"`
	Data         string  `json:"data"`
}

func (field *GeoLocationField) MarshalCSV() (string, error) {
	return field.String(), nil
}

func (field *GeoLocationField) UnmarshalCSV(csv string) error {
	parts := strings.Split(csv, ", ")
	if len(parts) != 2 {
		return errors.New("invalid CSV format for GeoLocationField")
	}

	lat, err := strconv.ParseFloat(parts[0], 32)
	if err != nil {
		return fmt.Errorf("invalid latitude value: %v", err)
	}

	lng, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return fmt.Errorf("invalid longitude value: %v", err)
	}

	field.Latitude = float32(lat)
	field.Longitude = float32(lng)

	return nil
}

func (field *GeoLocationField) String() string {
	return fmt.Sprintf("%g, %g", field.Latitude, field.Longitude)
}
