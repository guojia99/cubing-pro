package algdb

import (
	"testing"
)

func TestBldDB_Select(t *testing.T) {
	b := NewBldDB("/home/guojia/worker/code/cube/cubing-pro/build/alg/bld",
		"/home/guojia/worker/code/cube/cubing-pro/temp",
		"/home/guojia/worker/code/cube/cubing-pro/build/HuaWenHeiTi.ttf",
	)

	ls := []string{
		"bld 角[人造] PYR",
		"bld 角 PYR",
		"bld 角[圆子] PYR",
		"bld 角 UFR-UFL-URB",
		"bld 棱 UF-UL-UR",
	}

	for _, l := range ls {
		out, img, err := b.Select(l, nil)
		if err != nil {
			t.Error(err)
		}
		t.Log(out)
		t.Logf(img)
	}

}
