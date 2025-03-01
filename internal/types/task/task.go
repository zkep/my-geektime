package task

import "github.com/zkep/my-geektime/internal/types/geek"

type TaskListRequest struct {
	TaskPid     string `json:"task_pid"  form:"task_pid"`
	Keywords    string `json:"keywords"  form:"keywords"`
	Xstatus     int32  `json:"xstatus" form:"xstatus"`
	Tag         int32  `json:"tag"  form:"tag"`
	Direction   int32  `json:"direction"  form:"direction"`
	ProductType int32  `json:"product_type"  form:"product_type"`
	ProductForm int32  `json:"product_form"  form:"product_form"`
	Page        int    `json:"page" form:"page"`
	PerPage     int    `json:"perPage"  form:"perPage"`
}

type TaskListResponse struct {
	Count int64  `json:"count"`
	Rows  []Task `json:"rows"`
}

type Task struct {
	// task id
	TaskId string `json:"task_id,omitempty"`
	// task pid
	TaskPid string `json:"task_pid,omitempty"`
	// otherId
	OtherId string `json:"other_id,omitempty"`
	// task name
	TaskName string `json:"task_name,omitempty"`
	// task type
	TaskType string `json:"task_type,omitempty"`
	// cover
	Cover string `json:"cover,omitempty"`
	// status
	Status int32 `json:"status,omitempty"`
	// statistics
	Statistics TaskStatistics `json:"statistics,omitempty"`
	// Subtitle
	Subtitle string `json:"subtitle,omitempty"`
	// IntroHTML
	IntroHTML string `json:"intro_html,omitempty"`
	// dir
	Dir string `json:"dir,omitempty"`
	// doc
	Doc string `json:"doc,omitempty"`
	// object
	Object string `json:"object,omitempty"`
	// IsVideo
	IsVideo bool `json:"is_video,omitempty"`
	// IsAudio
	IsAudio bool `json:"is_audio,omitempty"`
	// sale
	Sale int `json:"sale,omitempty"`
	// SaleType
	SaleType int `json:"sale_type,omitempty"`
	// Share
	Share geek.ProductShare `json:"share,omitempty"`
	// Author
	Author geek.ArticleAuthor `json:"author,omitempty"`
	// Article
	Article geek.ProductArticle `json:"article,omitempty"`
	// created_at
	CreatedAt int64 `json:"created_at,omitempty"`
	// updated_at
	UpdatedAt int64 `json:"updated_at,omitempty"`
	// deleted_at
	DeletedAt int64 `json:"deleted_at,omitempty"`
}

type TaskStatistics struct {
	Count int         `json:"count"`
	Items map[int]int `json:"items"`
}

type TaskMessage struct {
	Object string `json:"object,omitempty"`
	Text   string `json:"text,omitempty"`
	Doc    string `json:"doc,omitempty"`
}

type RetryRequest struct {
	// task pid
	Pid string `json:"pid,omitempty" form:"pid" binding:"required"`
	// task ids
	Ids string `json:"ids,omitempty" form:"ids"`
	// retry
	Retry bool `json:"retry" form:"retry"`
}

type TaskInfoRequest struct {
	// task id
	Id string `json:"id,omitempty" form:"id" binding:"required"`
}

type TaskInfoResponse struct {
	// Task
	Task Task `json:"task"`
	// Article
	Article geek.ArticleInfo `json:"article"`
	// message
	Message TaskMessage `json:"message,omitempty"`
	// palyURL
	PalyURL string `json:"play_url,omitempty"`
}

type TaskDownloadRequest struct {
	// task id
	Id string `json:"id,omitempty" form:"id" binding:"required"`
	// type
	Type string `json:"type,omitempty" form:"type" binding:"required"`
	// url
	Url string `json:"url,omitempty" form:"url"`
}

type DeleteRequest struct {
	// task pid
	Pid string `json:"pid,omitempty" form:"pid"`
	// task ids
	Ids string `json:"ids,omitempty" form:"ids"`
}

type TaskKmsRequest struct {
	// task id
	Ciphertext string `json:"Ciphertext,omitempty" form:"Ciphertext" binding:"required"`
}

type TaskPlayRequest struct {
	// task id
	Id string `json:"id,omitempty" form:"id" binding:"required"`
}

type TaskPlayPartRequest struct {
	// part
	P string `json:"p,omitempty" form:"p" binding:"required"`
}

type TaskExportRequest struct {
	// task pid
	Pid string `json:"pid,omitempty" form:"pid" binding:"required"`
	// type
	Type string `json:"type,omitempty" form:"type" binding:"required"`
}
