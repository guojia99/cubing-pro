package algdb

import (
	"testing"
)

func TestBldDB_Select(t *testing.T) {
	b := NewBldDB("/home/guojia/worker/code/cube/cubing-pro/build/alg/bld")

	ls := []string{
		"bld 角[人造] PYR",
		"bld 角 PYR",
		"bld 角[圆子] PYR",
	}

	for _, l := range ls {
		out, _, err := b.Select(l, nil)
		if err != nil {
			t.Error(err)
		}
		t.Log(out)
	}

}
