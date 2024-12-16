package utils

import (
	"fmt"
	"time"
)

func PtrNow() *time.Time {
	t := time.Now()
	return &t

}

func PtrTime(t time.Time) *time.Time {
	return &t
}

func DurationToChinese(d time.Duration) string {
	if d <= 0 {
		return "0秒"
	}

	// 获取总秒数
	totalSeconds := int64(d.Seconds())
	seconds := totalSeconds % 60
	minutes := (totalSeconds / 60) % 60
	hours := (totalSeconds / 3600) % 24
	days := totalSeconds / (3600 * 24)

	result := ""
	if days > 0 {
		result += fmt.Sprintf("%d天", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%d小时", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%d分钟", minutes)
	}
	if seconds > 0 {
		result += fmt.Sprintf("%d秒", seconds)
	}

	return result
}
