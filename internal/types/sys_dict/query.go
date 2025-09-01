package sys_dict

import "github.com/zkep/my-geektime/internal/model"

type Query struct {
	Id int64 `uri:"id" form:"id" query:"id" json:"id"`
}

type DictTree struct {
	Data     []byte      `json:"-"`
	ID       int64       `json:"-"`
	Pkey     string      `json:"-"`
	Key      string      `json:"key"`
	Label    string      `json:"label"`
	Value    any         `json:"value"`
	Children []*DictTree `json:"children,omitempty"`
}

type DictOptions struct {
	Options []*DictTree `json:"options"`
}

type QueryTree struct {
	Key       string `form:"key"`
	FiledName string `form:"filedName"`
	NoChild   string `form:"noChild"`
}

type DictValue struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

const (
	CollectCategoryKey = "collectCategory"

	GeektimeCategoryKey = "geektimeCategory"
)

type DictNode struct {
	*model.SysDictBase
	Children []*DictNode
}
