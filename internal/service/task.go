package service

import (
	"github.com/google/uuid"
)

const (
	TASK_STATUS_PENDING  = 0x01
	TASK_STATUS_RUNNING  = 0x02
	TASK_STATUS_FINISHED = 0x03
	TASK_STATUS_ERROR    = 0x04
)

const (
	TASK_TYPE_PRODUCT = "product"
	TASK_TYPE_ARTICLE = "article"
)

func TaskID() string {
	return uuid.New().String()
}
