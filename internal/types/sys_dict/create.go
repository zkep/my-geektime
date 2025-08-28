package sys_dict

import "encoding/json"

type Request struct {
	Pkey    string          `json:"pkey"`
	Key     string          `json:"key"`
	Name    string          `json:"name"`
	Summary string          `json:"summary"`
	Content json.RawMessage `json:"content"`
	Sort    int32           `json:"sort"`
}

type Response struct {
	Id      int64 `json:"id"`
	Created int64 `json:"created"`
	Updated int64 `json:"updated"`
	Defer   bool  `json:"defer,omitempty"`
	Request
}

type UpdateDictRequest struct {
	Id int64 `json:"id"`
	Request
}
