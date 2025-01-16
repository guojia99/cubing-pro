package algdb

import (
	"fmt"
	"testing"
)

func TestCubePy_Select(t *testing.T) {

	s := NewCubePy("/home/guojia/worker/code/cube/cubing-pro/build/alg")
	out, _, err := s.Select("l4e S", nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out)
}
