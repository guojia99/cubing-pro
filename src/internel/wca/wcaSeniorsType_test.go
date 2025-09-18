package wca

import (
	"reflect"
	"testing"
)

func Test_getAges(t *testing.T) {
	t.Run("60", func(t *testing.T) {
		var want = []int{60, 50, 40}
		if got := getAges(60); reflect.DeepEqual(got, want) == false {
			t.Errorf("getAges() = %v, want %v", got, want)
		}
	})
}
