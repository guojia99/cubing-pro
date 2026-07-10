package wca

import (
	"encoding/json"
	"testing"
	"time"
)

func Test_wca_GetEventRankWithTimer(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, count, err := w.GetEventRankWithTimer("333", "China", 2023, 12, true, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("count: %d", count)

	dd, _ := json.MarshalIndent(out, "", "\t")
	t.Logf("out: %s", dd)
}

func Test_wca_GetEventRankWithFullNow(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, count, err := w.GetEventRankWithFullNow(
		"333mbf", "CN", true, 1, 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("count: %d", count)

	dd, _ := json.MarshalIndent(out, "", "\t")
	t.Logf("out: %s", dd)
}

func Test_wca_GetEventRankWithOnlyYear(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, count, err := w.GetEventRankWithOnlyYear(
		"333", "CN", 2019, 0, false, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("count: %d", count)

	dd, _ := json.MarshalIndent(out, "", "\t")
	t.Logf("out: %s", dd)
}

func Test_wca_GetPersonBestRanks(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, err := w.GetPersonBestRanks("2018SHEN07")
	if err != nil {
		t.Fatal(err)
	}

	dd, _ := json.MarshalIndent(out, "", "\t")
	t.Logf("out: %s", dd)
}

func TestGetMaxMonth(t *testing.T) {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	maxFinishedMonth := 12
	if currentMonth > 1 {
		maxFinishedMonth = currentMonth - 1
	}

	tests := []struct {
		name  string
		year  int
		month int
		want  int
	}{
		{
			name:  "今年 month=0",
			year:  currentYear,
			month: 0,
			want:  maxFinishedMonth,
		},
		{
			name:  "今年 month 小于最大月份",
			year:  currentYear,
			month: 3,
			want: func() int {
				if maxFinishedMonth >= 3 {
					return 3
				}
				return maxFinishedMonth
			}(),
		},
		{
			name:  "今年 month 等于最大月份",
			year:  currentYear,
			month: maxFinishedMonth,
			want:  maxFinishedMonth,
		},
		{
			name:  "今年 month 大于最大月份",
			year:  currentYear,
			month: 12,
			want:  maxFinishedMonth,
		},
		{
			name:  "去年 month=0",
			year:  currentYear - 1,
			month: 0,
			want:  12,
		},
		{
			name:  "去年 指定月份",
			year:  currentYear - 1,
			month: 5,
			want:  5,
		},
		{
			name:  "去年 month>12",
			year:  currentYear - 1,
			month: 15,
			want:  12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMaxMonth(tt.year, tt.month)
			if got != tt.want {
				t.Errorf("getMaxMonth(%d, %d) = %d, want %d",
					tt.year, tt.month, got, tt.want)
			}
		})
	}
}
