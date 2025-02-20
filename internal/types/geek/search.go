package geek

type SearchRequest struct {
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	Size     int    `json:"size"`
	Prev     int    `json:"prev"`
	Platform string `json:"platform"`
}

type SearchResponse struct {
	Code  int        `json:"code,omitempty"`
	Data  SearchData `json:"data,omitempty"`
	Error any        `json:"error,omitempty"`
}

type SearchData struct {
	List []SearchList `json:"list,omitempty"`
	Page Page         `json:"page,omitempty"`
}

type SearchList struct {
	ItemType string        `json:"item_type,omitempty"`
	Category string        `json:"category,omitempty"`
	Product  SearchProduct `json:"product,omitempty"`
	Score    int           `json:"score,omitempty"`
}

type SearchProduct struct {
	Title         string `json:"title,omitempty"`
	Subtitle      string `json:"subtitle,omitempty"`
	Cover         string `json:"cover,omitempty"`
	TotalLesson   int    `json:"total_lesson,omitempty"`
	VideoDuration string `json:"video_duration,omitempty"`
	Uint          string `json:"uint,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
	AuthorIntro   string `json:"author_intro,omitempty"`
	ID            int    `json:"id,omitempty"`
	Type          string `json:"type,omitempty"`
	Aid           int    `json:"aid,omitempty"`
	Subscribe     bool   `json:"subscribe,omitempty"`
}

type Page struct {
	More  bool `json:"more,omitempty"`
	Count int  `json:"count,omitempty"`
}
