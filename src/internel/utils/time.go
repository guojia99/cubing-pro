package utils

import (
	"time"
)

func PtrNow() *time.Time {
	t := time.Now()
	return &t

}

func PtrTime(t time.Time) *time.Time {
	return &t
}
