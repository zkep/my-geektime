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

var ALLStatus = []int{
	TASK_STATUS_PENDING,
	TASK_STATUS_RUNNING,
	TASK_STATUS_FINISHED,
	TASK_STATUS_ERROR,
}

func TaskID() string {
	src := uuid.New().String()
	slicedUUID := src[0:8] + src[9:13] + src[14:18] + src[19:23]
	return slicedUUID
}
