package geek

type DiscussionListRequest struct {
	UseLikesOrder bool  `json:"use_likes_order,omitempty" form:"use_likes_order"`
	TargetID      int64 `json:"target_id,omitempty" form:"target_id"`
	TargetType    int   `json:"target_type,omitempty" form:"target_type"`
	PageType      int   `json:"page_type,omitempty" form:"page_type"`
	Prev          int   `json:"prev,omitempty" form:"prev"`
	Size          int   `json:"size,omitempty" form:"size"`
	Page          int   `json:"-" form:"page"`
	PerPage       int   `json:"-"  form:"perPage"`
}

type DiscussionListResponse struct {
	Count int64            `json:"count"`
	Rows  []DiscussionData `json:"rows"`
}

type DiscussionOriginListResponse struct {
	Code  int            `json:"code,omitempty"`
	Data  DiscussionList `json:"data,omitempty"`
	Error any            `json:"error,omitempty"`
}

type DiscussionList struct {
	List []DiscussionData `json:"list,omitempty"`
	Page DiscussionPage   `json:"page,omitempty"`
}

type DiscussionData struct {
	Author                DiscussionAuthor   `json:"author,omitempty"`
	ReplyAuthor           ReplyAuthor        `json:"reply_author,omitempty"`
	Discussion            Discussion         `json:"discussion,omitempty"`
	Score                 int64              `json:"score,omitempty"`
	Extra                 string             `json:"extra,omitempty"`
	ChildDiscussionNumber int                `json:"child_discussion_number,omitempty"`
	ChildDiscussions      []ChildDiscussions `json:"child_discussions,omitempty"`
}

type DiscussionAuthor struct {
	ID        int    `json:"id,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Note      string `json:"note,omitempty"`
	Ucode     string `json:"ucode,omitempty"`
	RaceMedal int    `json:"race_medal,omitempty"`
	UserType  int    `json:"user_type,omitempty"`
	IsPvip    bool   `json:"is_pvip,omitempty"`
}

type ReplyAuthor struct {
	ID        int    `json:"id,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Note      string `json:"note,omitempty"`
	Ucode     string `json:"ucode,omitempty"`
	RaceMedal int    `json:"race_medal,omitempty"`
	UserType  int    `json:"user_type,omitempty"`
	IsPvip    bool   `json:"is_pvip,omitempty"`
}

type Discussion struct {
	ID                int64  `json:"id,omitempty"`
	DiscussionContent string `json:"discussion_content,omitempty"`
	LikesNumber       int64  `json:"likes_number,omitempty"`
	IsDelete          bool   `json:"is_delete,omitempty"`
	IsHidden          bool   `json:"is_hidden,omitempty"`
	Ctime             int64  `json:"ctime,omitempty"`
	IsLiked           bool   `json:"is_liked,omitempty"`
	CanDelete         bool   `json:"can_delete,omitempty"`
	IsComplain        bool   `json:"is_complain,omitempty"`
	IsTop             bool   `json:"is_top,omitempty"`
	ParentID          int    `json:"parent_id,omitempty"`
	IPAddress         string `json:"ip_address,omitempty"`
	GroupID           int    `json:"group_id,omitempty"`
}

type ChildDiscussions struct {
	Author      DiscussionAuthor `json:"author,omitempty"`
	ReplyAuthor ReplyAuthor      `json:"reply_author,omitempty"`
	Discussion  Discussion       `json:"discussion,omitempty"`
	Score       int              `json:"score,omitempty"`
	Extra       string           `json:"extra,omitempty"`
}

type DiscussionPage struct {
	Total   int64 `json:"total,omitempty"`
	More    bool  `json:"more,omitempty"`
	Current int64 `json:"current,omitempty"`
}
