package sys_dict

type ListRequest struct {
	Name    string `json:"name"    form:"name"`
	Key     string `json:"key"     form:"key"`
	Pkey    string `json:"pkey"    form:"pkey"`
	Page    int    `json:"page"    form:"page"`
	PerPage int    `json:"perPage" form:"perPage"`
}

type ListResponse struct {
	Count int64      `json:"count"`
	Rows  []Response `json:"rows"`
}

type ListItem struct {
	Response
	Children []*Response `json:"children"`
}
