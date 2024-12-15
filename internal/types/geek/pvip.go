package geek

const (
	SOURCE_FROM_ME   = "me"
	SOURCE_FROM_PVIP = "pvip"
)

type PvipProductRequest struct {
	TagIds       []any `json:"tag_ids"  form:"tag_ids"`
	ProductType  int   `json:"product_type"  form:"product_type"`
	ProductForm  int   `json:"product_form"  form:"product_form"`
	Pvip         int   `json:"pvip"  form:"pvip"`
	Prev         int   `json:"prev"  form:"prev"`
	Size         int   `json:"size"  form:"size"`
	Sort         int   `json:"sort"  form:"sort"`
	WithArticles bool  `json:"with_articles"  form:"with_articles"`
	Page         int   `json:"page,omitempty" form:"page"`
	PerPage      int   `json:"perPage,omitempty"  form:"perPage"`
}
