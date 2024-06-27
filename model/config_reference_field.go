package model

type ConfigReferenceField struct {
	TargetId   string `json:"target_id"`
	TargetType string `json:"target_type"`
	TargetUuid string `json:"target_uuid"`
}
