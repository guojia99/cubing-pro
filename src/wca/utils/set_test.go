package utils_tool

import "testing"

func TestHasIntersection(t *testing.T) {
	type args struct {
		a []string
		b []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				a: []string{"333", "444", "555"},
				b: []string{"444"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasIntersection(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("HasIntersection() = %v, want %v", got, tt.want)
			}
		})
	}
}
