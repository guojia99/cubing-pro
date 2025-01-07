package scramble

import (
	"testing"
)

func Test_scramble_WCACubeScramble(t *testing.T) {
	s := &scramble{
		endpoint: "http://localhost:2014",
	}

	type args struct {
		cube string
		nums int
	}
	tests := []struct {
		name string
	}{
		{"333"},
		{"333bf"},
		{"333oh"},
		{"333fm"},
		{"333mbf"},
		{"444bf"},
		{"555bf"},
		{"sq1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.WCACubeScramble(tt.name, 1)
			if err != nil {
				t.Errorf("WCACubeScramble() error = %v", err)
				return
			}
			t.Log(tt.name, got)
		})
	}
}

func Test_scramble_WCACubeScramble2(t *testing.T) {
	s := &scramble{
		endpoint: "http://localhost:2014",
	}
	var tests []struct {
		name string
	}

	for _, a := range s.events() {
		tests = append(tests, struct{ name string }{a.ID})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.WCACubeScramble(tt.name, 1)
			if err != nil {
				t.Errorf("WCACubeScramble() error = %v", err)
				return
			}
			t.Log(tt.name, got)
		})
	}
}
