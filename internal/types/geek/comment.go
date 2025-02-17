package geek

type ArticleCommentListRequest struct {
	Aid     int64 `json:"aid" form:"aid" binding:"required"`
	Prev    int   `json:"prev" form:"prev"`
	Sort    int   `json:"order" form:"order"`
	Page    int   `json:"-" form:"page"`
	PerPage int   `json:"-"  form:"perPage"`
}

type ArticleCommentListResponse struct {
	Count int64            `json:"count"`
	Rows  []ArticleComment `json:"rows"`
}

type ArticleCommentList struct {
	Code  int                `json:"code,omitempty"`
	Data  ArticleCommentData `json:"data,omitempty"`
	Error any                `json:"error,omitempty"`
}

type ArticleCommentData struct {
	List []ArticleComment `json:"list,omitempty"`
	Page GeekPage         `json:"page,omitempty"`
}

type GeekPage struct {
	Count int64 `json:"count,omitempty"`
	More  bool  `json:"more,omitempty"`
}

type ArticleComment struct {
	HadLiked        bool                    `json:"had_liked,omitempty"`
	ID              int64                   `json:"id,omitempty"`
	UserName        string                  `json:"user_name,omitempty"`
	CanDelete       bool                    `json:"can_delete,omitempty"`
	ProductType     string                  `json:"product_type,omitempty"`
	UID             int                     `json:"uid,omitempty"`
	IPAddress       string                  `json:"ip_address,omitempty"`
	Ucode           string                  `json:"ucode,omitempty"`
	UserHeader      string                  `json:"user_header,omitempty"`
	CommentIsTop    bool                    `json:"comment_is_top,omitempty"`
	CommentCtime    int64                   `json:"comment_ctime,omitempty"`
	IsPvip          bool                    `json:"is_pvip,omitempty"`
	Replies         []ArticleCommentReplies `json:"replies,omitempty"`
	DiscussionCount int64                   `json:"discussion_count,omitempty"`
	RaceMedal       int                     `json:"race_medal,omitempty"`
	Score           int                     `json:"score,omitempty"`
	ProductID       int                     `json:"product_id,omitempty"`
	CommentContent  string                  `json:"comment_content,omitempty"`
	LikeCount       int64                   `json:"like_count,omitempty"`
}

type ArticleCommentReplies struct {
	ID           int    `json:"id,omitempty"`
	Content      string `json:"content,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	UserNameReal string `json:"user_name_real,omitempty"`
	UID          int    `json:"uid,omitempty"`
	Ctime        int    `json:"ctime,omitempty"`
	IPAddress    string `json:"ip_address,omitempty"`
	CommentID    int    `json:"comment_id,omitempty"`
	Utype        int    `json:"utype,omitempty"`
}
