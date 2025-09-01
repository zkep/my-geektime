package collect

import "encoding/json"

type CreateRequest struct {
	Ids         string      `json:"ids,omitempty" form:"ids"`
	CollectType string      `json:"collect_type,omitempty" form:"collect_type"`
	Category    json.Number `json:"category,omitempty" form:"category"`
}

const (
	CollectTask = "task"
)
