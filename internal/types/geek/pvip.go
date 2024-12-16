package geek

type PvipProductRequest struct {
	TagIds       []int `json:"tag_ids"  form:"tag_ids"`
	ProductType  int   `json:"product_type"  form:"product_type"`
	ProductForm  int   `json:"product_form"  form:"product_form"`
	Pvip         int   `json:"pvip"  form:"pvip"`
	Prev         int   `json:"prev"  form:"prev"`
	Size         int   `json:"size"  form:"size"`
	Sort         int   `json:"sort"  form:"sort"`
	WithArticles bool  `json:"with_articles"  form:"with_articles"`
	Direction    int   `json:"-"  form:"direction"`
	Tag          int   `json:"-"  form:"tag"`
	Page         int   `json:"-" form:"page"`
	PerPage      int   `json:"-"  form:"perPage"`
}
