package model

type EntityReferenceField struct {
	TargetId   int    `json:"target_id"`
	TargetType string `json:"target_type"`
	TargetUuid string `json:"target_uuid"`
	Url        string `json:"url"`
}
