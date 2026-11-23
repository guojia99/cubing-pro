package siteannouncement

import (
	"strings"
	"testing"
)

func TestNormalizeText(t *testing.T) {
	for _, tc := range []struct {
		text string
		want bool
	}{
		{" \n\t ", false},
		{" 公告  ", true},
		{strings.Repeat("魔", 300), true},
		{strings.Repeat("🧩", 301), false},
	} {
		input := content{Text: tc.text}
		if err := input.normalize(); (err == nil) != tc.want || input.Text != strings.TrimSpace(tc.text) {
			t.Errorf("unexpected result for %d bytes: %v", len(tc.text), err)
		}
	}
}

func TestValidLink(t *testing.T) {
	for _, tc := range []struct {
		link string
		want bool
	}{
		{"", true},
		{"/updates?topic=site#top", true},
		{"https://example.com/updates", true},
		{"//example.com", false},
		{`/\example.com`, false},
		{"/\texample.com", false},
		{"/%", false},
		{"http://example.com", false},
		{"https://", false},
		{"https://user@example.com", false},
		{"javascript:void(0)", false},
		{"relative/updates", false},
	} {
		if got := validLink(tc.link); got != tc.want {
			t.Errorf("validLink(%q) = %t, want %t", tc.link, got, tc.want)
		}
	}
}
