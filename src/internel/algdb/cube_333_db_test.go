package algdb

import (
	"fmt"
	"testing"
)

func TestNewCube333(t *testing.T) {
	c := NewCube333("/home/guojia/worker/code/cube/cubing-pro/build/alg")
	for k, _ := range c.pll.Alg {
		fmt.Println(k)
	}
}
