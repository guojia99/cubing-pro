package wca

import (
	"fmt"
	"testing"
)

func Test_decodeEvents(t *testing.T) {
	type args struct {
		mask uint64
	}
	tests := []struct {
		args args
	}{
		{
			args{mask: 16383},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.args.mask), func(t *testing.T) {
			got := decodeEvents(tt.args.mask)
			t.Logf("%v", got)
		})
	}
}
