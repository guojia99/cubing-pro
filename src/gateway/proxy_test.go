package gateway

import (
	"net/http"
	"testing"
)

func TestNormalizeProxiedResponse_charset(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Content-Type", "text/html")
	resp.Header.Set("Content-Encoding", "gzip")
	resp.Header.Set("Content-Length", "123")

	if err := normalizeProxiedResponse(resp); err != nil {
		t.Fatal(err)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	if resp.Header.Get("Content-Encoding") != "" {
		t.Fatalf("Content-Encoding should be stripped")
	}
	if resp.Header.Get("Content-Length") != "" {
		t.Fatalf("Content-Length should be stripped")
	}
}

func TestNormalizeProxiedResponse_skipExistingCharset(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Content-Type", "text/html; charset=gbk")

	if err := normalizeProxiedResponse(resp); err != nil {
		t.Fatal(err)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/html; charset=gbk" {
		t.Fatalf("Content-Type = %q, should not change existing charset", got)
	}
}

func TestContentTypeWithUTF8(t *testing.T) {
	if got := contentTypeWithUTF8("/app/index.html"); got != "text/html; charset=utf-8" {
		t.Fatalf("html: got %q", got)
	}
	if got := contentTypeWithUTF8("/app/bundle.js"); got != "application/javascript; charset=utf-8" {
		t.Fatalf("js: got %q", got)
	}
	if got := contentTypeWithUTF8("/app/logo.png"); got != "" {
		t.Fatalf("png: got %q, want empty", got)
	}
}
