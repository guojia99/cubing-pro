package gateway

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/configs"
)

func TestServeDefaultStatic_nextExport(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "index.html"), "<html>home</html>")
	algsDir := filepath.Join(root, "algs", "222", "EG")
	if err := os.MkdirAll(algsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(algsDir, "index.html"), "<html>eg</html>")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/algs/222/EG", nil)

	serveDefaultStatic(ctx, configs.GatewayConfig{
		StaticRoot: root,
		SPA:        false,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); body != "<html>eg</html>" {
		t.Fatalf("body = %q, want eg page", body)
	}
}

func TestServeDefaultStatic_legacySPA(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	root := t.TempDir()
	index := filepath.Join(root, "index.html")
	staticDir := filepath.Join(root, "static")
	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, index, "<html>spa</html>")
	writeFile(t, filepath.Join(staticDir, "app.js"), "console.log('ok')")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/welcome", nil)

	serveDefaultStatic(ctx, configs.GatewayConfig{
		IndexPath:  index,
		StaticPath: staticDir,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); body != "<html>spa</html>" {
		t.Fatalf("body = %q, want spa shell", body)
	}
}

func TestServeDefaultStatic_dynamicRouteFallback(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "index.html"), "<html>home</html>")
	dynamicDir := filepath.Join(root, "competition", "__dynamic__")
	if err := os.MkdirAll(dynamicDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dynamicDir, "index.html"), "<html>competition shell</html>")

	fallbacks := []configs.DynamicRouteFallbackConfig{
		{
			Match:       `^/competition/[^/]+/?$`,
			Placeholder: "/competition/__dynamic__/index.html",
		},
	}

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/competition/365", nil)

	serveDefaultStatic(ctx, configs.GatewayConfig{
		StaticRoot:            root,
		SPA:                   false,
		DynamicRouteFallbacks: fallbacks,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); body != "<html>competition shell</html>" {
		t.Fatalf("body = %q, want competition shell", body)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
