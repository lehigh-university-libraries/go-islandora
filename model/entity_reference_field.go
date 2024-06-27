package model

import "strconv"

type EntityReferenceField struct {
	TargetId   int    `json:"target_id"`
	TargetType string `json:"target_type"`
	TargetUuid string `json:"target_uuid"`
	Url        string `json:"url"`
}

func (field *EntityReferenceField) MarshalCSV() (string, error) {
	return strconv.Itoa(field.TargetId), nil
}

func (field *EntityReferenceField) UnmarshalCSV(csv string) error {
	var err error
	field.TargetId, err = strconv.Atoi(csv)

	return err
}
func (field *EntityReferenceField) String() string {
	return strconv.Itoa(field.TargetId)
}
