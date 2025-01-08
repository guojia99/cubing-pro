package script

import (
	"testing"
)

func TestCommutator(t *testing.T) {
	InitCommutator("./commutator.js")

	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{
			name: "JAG",
			args: "R F R' U R' D' R U' R' D R2 F' R'",
			want: "R F R':[U,R' D' R]",
		},
		{
			name: "JAG-2",
			args: "x' R2 D2 R' U' R D2 R' U R' x",
			want: "x' R:[R D2 R',U']",
		},
		{
			name:    "JDG",
			args:    "l' U R' D2 R U' R' D2 R2 x'",
			wantErr: true,
		},
		{
			name: "AEW",
			args: "L' U S' L2 S L2 U' L",
			want: "L' U:[S',L2]",
		},
		{
			name:    "AEW-2",
			args:    "r' F E' r2 E' r2 F' r",
			wantErr: true,
		},
		{
			name: "CW-CCW",
			args: "U' R D R' D' R D R' U R D' R' D R D' R'",
			want: "[U',R D R' D' R D R']",
		},
		{
			name: "X-JAC",
			args: "U' r' u' r U' r' u r U2",
			want: "U':[r' u' r,U']",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Commutator(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Commutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Commutator() got = %v, want %v", got, tt.want)
			}
			t.Logf("Commutator() [%s] alg =%v got = %v", tt.name, tt.args, got)
		})
	}
}
