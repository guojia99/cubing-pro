package gateway

import (
	"mime"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strings"
)

// utf8TextMediaTypes 需在 Content-Type 中显式声明 charset=utf-8 的文本类型（Safari/iOS 对缺省编码较敏感）。
var utf8TextMediaTypes = map[string]struct{}{
	"text/html":              {},
	"text/css":               {},
	"text/javascript":        {},
	"application/javascript": {},
	"application/json":       {},
	"text/plain":             {},
	"application/xml":        {},
	"text/xml":               {},
}

// configureReverseProxy 统一处理反向代理与 gateway gzip 中间件的配合：
// 1. 不向 upstream 请求压缩，避免 upstream gzip + gin gzip 双重压缩（iOS Safari 易把乱码当 JS 解析报语法错）
// 2. 去掉 upstream 的 Content-Encoding，由 gin gzip 对客户端只压缩一次
// 3. 为文本响应补全 charset=utf-8
func configureReverseProxy(p *httputil.ReverseProxy) {
	origDirector := p.Director
	p.Director = func(req *http.Request) {
		if origDirector != nil {
			origDirector(req)
		}
		req.Header.Del("Accept-Encoding")
	}

	origModify := p.ModifyResponse
	p.ModifyResponse = func(resp *http.Response) error {
		if origModify != nil {
			if err := origModify(resp); err != nil {
				return err
			}
		}
		return normalizeProxiedResponse(resp)
	}
}

func normalizeProxiedResponse(resp *http.Response) error {
	resp.Header.Del("Content-Encoding")
	resp.Header.Del("Content-Length")

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		return nil
	}
	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil
	}
	if _, ok := utf8TextMediaTypes[mediaType]; !ok {
		return nil
	}
	if _, ok := params["charset"]; ok {
		return nil
	}
	resp.Header.Set("Content-Type", mediaType+"; charset=utf-8")
	return nil
}

// contentTypeWithUTF8 为本地静态文件补全 charset（在 ctx.File 之前设置）。
func contentTypeWithUTF8(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml; charset=utf-8"
	default:
		return ""
	}
}
