package scramble

import (
	"os"
	"testing"
)

func Test_scramble_SImageWith2mf81(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{
			name: "minx",
			str:  "R++ D++ R++ D-- R-- D-- R-- D++ R-- D-- U'\nR++ D++ R++ D++ R-- D-- R++ D-- R++ D++ U\nR-- D++ R-- D-- R-- D-- R-- D++ R++ D++ U\nR-- D-- R-- D++ R++ D++ R++ D-- R-- D-- U'\nR++ D++ R++ D++ R-- D++ R++ D++ R-- D-- U'\nR-- D-- R-- D++ R-- D++ R++ D++ R-- D++ U\nR++ D-- R-- D++ R++ D++ R++ D-- R-- D-- U'",
		},
		{
			name: "pyram",
			str:  "L' R B U' B R' U R U R' U' r b'",
		},
		{
			name: "skewb",
			str:  "U R B U L' R' B U' B R' U'",
		},
		{
			name: "clock",
			str:  "UR0+ DR5+ DL4+ UL1+ U2- R2- D2- L2+ ALL5+ y2 U1- R4+ D2- L5- ALL4-",
		},
		{
			name: "sq-1",
			str:  "(-2, 0) / (3, -3) / (-3, 0) / (-1, -1) / (-3, 0) / (0, -3) / (3, 0) / (-3, 0) / (1, 0) / (-3, -3) / (-1, -2) / (2, -4) / (4, 0)",
		},
		{
			name: "777",
			str:  "U",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &scramble{
				scrambleDrawType: scrambleTypeDrawType2Mf8,
			}

			got, err := s.SImageWith2mf8(tt.str, tt.name)
			if err != nil {
				t.Error(err)
				return
			}
			_ = os.WriteFile(tt.name+".jpg", []byte(got), 0644)
		})
	}
}
