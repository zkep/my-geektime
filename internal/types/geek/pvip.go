package geek

type PvipProductRequest struct {
	TagIds       []int32 `json:"tag_ids"  form:"tag_ids"`
	ProductType  int32   `json:"product_type"  form:"product_type"`
	ProductForm  int32   `json:"product_form"  form:"product_form"`
	Pvip         int32   `json:"pvip"  form:"pvip"`
	Sort         int32   `json:"sort"  form:"sort"`
	WithArticles bool    `json:"with_articles"  form:"with_articles"`
	Prev         int     `json:"prev"  form:"prev"`
	Size         int     `json:"size"  form:"size"`
	Tag          int32   `json:"-"  form:"tag"`
	Direction    int32   `json:"-"  form:"direction"`
	Page         int     `json:"-" form:"page"`
	PerPage      int     `json:"-"  form:"perPage"`
}
