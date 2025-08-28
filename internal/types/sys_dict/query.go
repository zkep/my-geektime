package sys_dict

type Query struct {
	Id int64 `uri:"id" form:"id" query:"id" json:"id"`
}

type DictTree struct {
	ID       int64       `json:"-"`
	Pkey     string      `json:"-"`
	Key      string      `json:"-"`
	Label    string      `json:"label"`
	Value    any         `json:"value"`
	Children []*DictTree `json:"children,omitempty"`
}

type DictOptions struct {
	Options []*DictTree `json:"options"`
}

type QueryWithKey struct {
	Key    string `uri:"key" form:"key" query:"key"`
	Option string `uri:"option" form:"option" query:"option"`
}

type DictValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
