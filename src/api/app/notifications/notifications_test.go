package notifications

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestMergedNotificationWindow(t *testing.T) {
	base := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
	announcements := []notification{
		{ID: "site-announcement:a", CreatedAt: base.Add(4 * time.Minute)},
		{ID: "site-announcement:b", CreatedAt: base.Add(11 * time.Minute)},
		{ID: "site-announcement:c", CreatedAt: base.Add(-time.Minute)},
	}
	replies := []notification{}
	for i := 1; i <= 29; i++ {
		replies = append(replies, notification{ID: fmt.Sprintf("formula-reply:%d", i), CreatedAt: base.Add(time.Duration(i/2) * time.Minute)})
	}
	sort.Slice(replies, func(i, j int) bool {
		if replies[i].CreatedAt.Equal(replies[j].CreatedAt) {
			return replies[i].ID > replies[j].ID
		}
		return replies[i].CreatedAt.After(replies[j].CreatedAt)
	})
	all := append(append([]notification{}, announcements...), replies...)
	expected := notificationWindow(all, 0, len(all))
	for _, size := range []int{1, 5, 20} {
		for start := 0; start < len(all); start += size {
			offset := max(0, start-len(announcements))
			window := append(append([]notification{}, announcements...), replies[offset:min(offset+size+len(announcements), len(replies))]...)
			actual := notificationWindow(window, start-offset, size)
			if want := expected[start:min(start+size, len(all))]; !reflect.DeepEqual(actual, want) {
				t.Fatalf("window start=%d size=%d: got %#v, want %#v", start, size, actual, want)
			}
		}
	}
}
