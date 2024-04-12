package utils

import (
	"testing"
)

func TestGetID(t *testing.T) {
	tests := []struct {
		baseName string
	}{
		{baseName: "徐永浩"},
		{baseName: "孙大圣"},
		{baseName: "小丫鬟"},
		{baseName: "MIT-B"},
		{baseName: "熙-源~"},
		{baseName: "嘉吖"},
		{baseName: "mmmm"},
		{baseName: "徐子怡"},
	}
	for _, tt := range tests {
		t.Run(
			tt.baseName, func(t *testing.T) {
				got := GetIDButNotNumber(tt.baseName)
				t.Logf("%s \t %s", tt.baseName, got)
			},
		)
	}
}
