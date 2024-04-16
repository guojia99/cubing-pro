package scramble

import (
	"encoding/json"
	"testing"
)

func Test_scramble_get(t *testing.T) {
	s := &scramble{}
	s.url = "http://127.0.0.1:2014"

}

func Test_scramble_CubeScramble(t *testing.T) {
	url := "http://127.0.0.1:2014"
	type args struct {
		cube string
		nums int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1-333",
			args: args{
				cube: "333",
				nums: 1,
			},
		},
		{
			name: "3-333",
			args: args{
				cube: "333",
				nums: 3,
			},
		},
		{
			name: "3-333bf",
			args: args{
				cube: "333bf",
				nums: 3,
			},
		},
		{
			name: "3-sq1",
			args: args{
				cube: "sq1",
				nums: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := scramble{url: url}
				got, err := s.CubeScramble(tt.args.cube, tt.args.nums)
				if err != nil {
					t.Fatal(err)
				}
				d, _ := json.Marshal(got)
				t.Log(string(d))
			},
		)
	}
}
