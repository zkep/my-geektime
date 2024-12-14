package task

import "github.com/zkep/mygeektime/internal/types/geek"

type TaskListRequest struct {
	TaskPid string `json:"task_pid"  form:"task_pid"`
	Xstatus int    `json:"xstatus" form:"xstatus"`
	Page    int    `json:"page" form:"page"`
	PerPage int    `json:"perPage"  form:"perPage"`
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
	// status
	Status int32 `json:"status,omitempty"`
	// statistics
	Statistics TaskStatistics `json:"statistics,omitempty"`
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
	Object string `json:"object"`
	Text   string `json:"text"`
}

type RetryRequest struct {
	// task pid
	Pid string `json:"pid,omitempty" form:"pid" binding:"required"`
	// task ids
	Ids string `json:"ids,omitempty" form:"ids" binding:"required"`
}

type TaskInfoRequest struct {
	// task id
	Id string `json:"id,omitempty" form:"id" binding:"required"`
}

type TaskInfoResponse struct {
	// Task
	Task
	// ArticleInfo
	geek.ArticleInfo
	// message
	Message TaskMessage `json:"message,omitempty"`
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
