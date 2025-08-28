package collect

import (
	"encoding/json"

	"github.com/zkep/my-geektime/internal/model"
)

type CollectListRequest struct {
	Page     int    `json:"page" form:"page"`
	PerPage  int    `json:"perPage"  form:"perPage"`
	Category string `json:"category"  form:"category"`
}

type CollectListResponse struct {
	Count int64     `json:"count"`
	Rows  []Collect `json:"rows"`
}

type Collect struct {
	*model.Collect
	Item json.RawMessage `json:"item"`
}
