package utils

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	type args struct {
		s   string
		sep string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "vs",
			args: args{
				s:   " Max Park VS Feliks Zemdegs ",
				sep: "VS",
			},
			want: []string{
				"Max Park", "Feliks Zemdegs",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Split(tt.args.s, tt.args.sep); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Split() = %v, want %v", got, tt.want)
			}
		})
	}
}
