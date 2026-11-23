package siteannouncement

import (
	"strings"
	"testing"
)

func TestNormalizeContent(t *testing.T) {
	for _, tc := range []struct {
		name, title, text string
		want              bool
	}{
		{"missing title", "", "正文", false},
		{"blank title", " \n\t ", "正文", false},
		{"missing body", "标题", "", false},
		{"blank body", "标题", " \n\t ", false},
		{"minimum", " 魔 ", " 🧩 ", true},
		{"maximum", " " + strings.Repeat("🧩", 80) + " ", "\n" + strings.Repeat("魔", 1500) + "\n", true},
		{"title too long", strings.Repeat("🧩", 81), "正文", false},
		{"body too long", "标题", strings.Repeat("🧩", 1501), false},
		{"internal newlines", " 标题 ", " \n第一行\n\n  第二行\r\n第三行\n ", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := content{Title: tc.title, Text: tc.text}
			if err := input.normalize(); (err == nil) != tc.want {
				t.Fatalf("normalize() error = %v, want valid %t", err, tc.want)
			}
			if input.Title != strings.TrimSpace(tc.title) || input.Text != strings.TrimSpace(tc.text) {
				t.Fatalf("normalization changed internal content: %+v", input)
			}
		})
	}
}
