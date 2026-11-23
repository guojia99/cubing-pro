package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	announcementPublic = "/public/site-announcement"
	announcementAdmin  = "/admin/system_result/site-announcement"
)

type announcementResponse struct {
	Active   bool   `json:"active"`
	Revision string `json:"revision"`
	Text     string `json:"text"`
	Href     string `json:"href"`
}

func TestSiteAnnouncementRoutes(t *testing.T) {
	var dialector gorm.Dialector = sqlite.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	// Only point this at an empty, disposable test DB with permission to create tables.
	if dsn := os.Getenv("CUBING_ANNOUNCEMENT_TEST_MYSQL_DSN"); dsn != "" {
		dialector = mysql.Open(dsn)
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&system.KeyValue{}, &user.User{}); err != nil {
		t.Fatal(err)
	}
	service := &svc.Svc{DB: db}
	jwt := middleware.InitJWT(service)
	middleware.InitCheckAuth(service)
	router := gin.New()
	PublicRouters(router.Group(""), service)
	AdminRouters(router.Group(""), service)

	tokenFor := func(role user.Auth, name string) string {
		t.Helper()
		u := user.User{Name: name, CubeID: name, Auth: role}
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
		token, _, err := jwt.TokenGenerator(middleware.JwtMapClaims{Id: u.ID, Auth: role})
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	admin := tokenFor(user.AuthSuperAdmin, "announcement-superadmin")
	ordinary := tokenFor(user.AuthAdmin, "announcement-admin")

	request := func(method, path, body, token string, wantCode int) json.RawMessage {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			r.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != wantCode {
			t.Fatalf("%s %s: status=%d, want %d; %s", method, path, w.Code, wantCode, w.Body)
		}
		if wantCode != http.StatusOK {
			return nil
		}
		if method == http.MethodGet && w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: missing no-store", path)
		}
		var envelope struct {
			Code string          `json:"code"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil || envelope.Code != "200" || len(envelope.Data) == 0 {
			t.Fatalf("invalid response envelope: %s (%v)", w.Body, err)
		}
		return envelope.Data
	}
	adminRequest := func(method, body string) announcementResponse {
		t.Helper()
		raw := request(method, announcementAdmin, body, admin, http.StatusOK)
		var out announcementResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	expectPublic := func(want *announcementResponse) {
		t.Helper()
		raw := request(http.MethodGet, announcementPublic, "", "", http.StatusOK)
		if want == nil {
			if string(raw) != "null" {
				t.Fatalf("expected a null public announcement, got %s", raw)
			}
			return
		}
		var fields map[string]string
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		if len(fields) != 3 || fields["revision"] != want.Revision || fields["text"] != want.Text || fields["href"] != want.Href {
			t.Fatalf("unexpected public fields: %s", raw)
		}
	}

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		request(method, announcementAdmin, "", "", http.StatusUnauthorized)
		request(method, announcementAdmin, "", ordinary, http.StatusUnauthorized)
	}
	expectPublic(nil)
	if got := adminRequest(http.MethodGet, ""); got != (announcementResponse{}) {
		t.Fatalf("unexpected initial config: %+v", got)
	}
	if got := adminRequest(http.MethodDelete, ""); got != (announcementResponse{}) {
		t.Fatalf("withdrawing unused slot unexpectedly changed it: %+v", got)
	}
	request(http.MethodPut, announcementAdmin, `{"text":"   "}`, admin, http.StatusUnprocessableEntity)
	request(http.MethodPut, announcementAdmin, `{"text":"公告","href":"http://example.com/"}`, admin, http.StatusUnprocessableEntity)
	request(http.MethodPut, announcementAdmin, `{"text":`, admin, http.StatusBadRequest)

	first := adminRequest(http.MethodPut, `{"text":"  站点维护  ","href":" /welcome "}`)
	if !first.Active || first.Revision == "" || first.Text != "站点维护" || first.Href != "/welcome" {
		t.Fatalf("unexpected first publish: %+v", first)
	}
	expectPublic(&first)
	if got := adminRequest(http.MethodPut, `{"text":"站点维护","href":"/welcome"}`); got != first {
		t.Fatalf("no-op save changed revision: %+v", got)
	}

	edited := adminRequest(http.MethodPut, `{"text":"站点维护（时间更新）","href":"/welcome"}`)
	if !edited.Active || edited.Revision == first.Revision || edited.Text != "站点维护（时间更新）" {
		t.Fatalf("text edit failed to update revision: %+v", edited)
	}
	expectPublic(&edited)
	linked := adminRequest(http.MethodPut, `{"text":"站点维护（时间更新）","href":"https://example.com/updates"}`)
	if !linked.Active || linked.Revision == edited.Revision || linked.Text != edited.Text || linked.Href == edited.Href {
		t.Fatalf("link-only edit failed to update revision: %+v", linked)
	}
	expectPublic(&linked)

	withdrawn := adminRequest(http.MethodDelete, "")
	linked.Active = false
	if withdrawn != linked {
		t.Fatalf("withdrawal did not preserve content: got %+v, want %+v", withdrawn, linked)
	}
	expectPublic(nil)

	republished := adminRequest(http.MethodPut, `{"text":"站点维护（时间更新）","href":"https://example.com/updates"}`)
	if !republished.Active || republished.Revision == withdrawn.Revision || republished.Text != withdrawn.Text || republished.Href != withdrawn.Href {
		t.Fatalf("republish did not create a revision: %+v", republished)
	}
	expectPublic(&republished)
}
