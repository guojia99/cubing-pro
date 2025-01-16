package algdb

import (
	"testing"
)

func TestNewCube222(t *testing.T) {
	NewCube222("/home/guojia/worker/code/cube/cubing-pro/build/alg")
}

func TestCube222_Select(t *testing.T) {

	tests := []struct {
		args string
	}{
		{"222 eg1"},
		{"222 cll s1"},
		{"222 eg0 as4"},
		{"222 eg0 as8"},
		{"222 eg2 s1"},
	}
	c := NewCube222("/home/guojia/worker/code/cube/cubing-pro/build/alg")
	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			out, img, err := c.Select(tt.args, nil)
			if err != nil {
				t.Fatalf("cube222.Select() error = %v", err)
			}
			t.Log("got out:", out)
			t.Log("got img:", img)
		})
	}
}
