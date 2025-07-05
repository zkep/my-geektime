package utils

import "github.com/google/uuid"

func HalfUUID() string {
	src := uuid.New().String()
	slicedUUID := src[0:8] + src[9:13] + src[14:18] + src[19:23]
	return slicedUUID
}
