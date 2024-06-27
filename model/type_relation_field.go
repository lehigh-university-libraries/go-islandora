package model

type TypedRelationField struct {
	TargetId int    `json:"target_id"`
	RelType  string `json:"rel_type"`
}
