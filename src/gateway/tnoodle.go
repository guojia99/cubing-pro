package gateway

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func (g *Gateway) runTNoodleApi() {
	if g.cfg.Gateway.OutsizeTNoodlePort == 0 {
		return
	}
	log.Printf("start tNoodle api %d -> %d\n", g.cfg.Gateway.OutsizeTNoodlePort, g.cfg.Gateway.TNoodlePort)

	g.tNoodleApi = gin.Default()
	g.tNoodleApi.NoRoute(g.tNoodleRoute())
	err := g.tNoodleApi.Run(fmt.Sprintf(":%d", g.cfg.Gateway.OutsizeTNoodlePort))
	if err != nil {
		log.Fatalf("failed to start tNoodle api: %v", err)
	}
}

func (g *Gateway) tNoodleRoute() gin.HandlerFunc {
	api, _ := url.Parse(fmt.Sprintf("http://localhost:%d", g.cfg.Gateway.TNoodlePort))

	baseProxy := httputil.NewSingleHostReverseProxy(api)
	jsProxy := httputil.NewSingleHostReverseProxy(api)

	jsProxy.ModifyResponse = func(resp *http.Response) error {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()

		newBody := bytes.ReplaceAll(body,
			[]byte("http://localhost:2014"),
			[]byte("http://cubing.pro:12014"))

		resp.Body = io.NopCloser(bytes.NewReader(newBody))
		resp.ContentLength = int64(len(newBody))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
		resp.Header.Del("Content-Encoding") // 防止 gzip 干扰
		return nil
	}

	return func(ctx *gin.Context) {
		// 只有 .js 文件才需要替换内容
		if strings.HasSuffix(ctx.Request.URL.Path, ".js") {
			jsProxy.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}
		baseProxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
