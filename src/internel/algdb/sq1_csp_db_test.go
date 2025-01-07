package algdb

import (
	"fmt"
	"testing"
)

func TestSQ1CspDB_getImage(t *testing.T) {
	base := cspAlg{
		Image: []string{
			"8.jpg",
			"star.jpg",
		},
	}

	mirror := cspAlg{
		Image: []string{
			"m_star.jpg",
			"m_8.jpg",
		},
	}
	s := NewSQ1CspDB(
		"/home/guojia/worker/code/cube/cubing-pro/build/alg/csp.json",
		"/home/guojia/worker/code/cube/cubing-pro/build/alg/csp_image",
		"/home/guojia/worker/code/cube/cubing-pro/temp",
		"/home/guojia/worker/code/cube/cubing-pro/build/HuaWenHeiTi.ttf",
	)
	out := s.getImage(base, mirror)
	fmt.Println(out)
}
