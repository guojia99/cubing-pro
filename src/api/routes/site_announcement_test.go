package routes

import (
	"encoding/json"
	"fmt"
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
	ID       string `json:"id"`
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
	adminMutation := func(method, id, body string) announcementResponse {
		t.Helper()
		path := announcementAdmin
		if id != "" {
			path += "/" + id
		}
		raw := request(method, path, body, admin, http.StatusOK)
		var out announcementResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	adminList := func() []announcementResponse {
		t.Helper()
		raw := request(http.MethodGet, announcementAdmin, "", admin, http.StatusOK)
		var list []announcementResponse
		if err := json.Unmarshal(raw, &list); err != nil || list == nil {
			t.Fatalf("expected an admin array, got %s (%v)", raw, err)
		}
		return list
	}
	expectPublic := func(want ...announcementResponse) {
		t.Helper()
		raw := request(http.MethodGet, announcementPublic, "", "", http.StatusOK)
		var fields []map[string]string
		if err := json.Unmarshal(raw, &fields); err != nil || fields == nil || len(fields) != len(want) {
			t.Fatalf("expected %d public announcements, got %s (%v)", len(want), raw, err)
		}
		for i, entry := range fields {
			if len(entry) != 4 || entry["id"] != want[i].ID || entry["revision"] != want[i].Revision || entry["text"] != want[i].Text || entry["href"] != want[i].Href {
				t.Fatalf("unexpected public fields: %s", raw)
			}
		}
	}

	for _, tc := range []struct{ method, suffix string }{
		{http.MethodGet, ""}, {http.MethodPost, ""},
		{http.MethodPut, "/some-id"}, {http.MethodDelete, "/some-id"},
	} {
		request(tc.method, announcementAdmin+tc.suffix, "", "", http.StatusUnauthorized)
		request(tc.method, announcementAdmin+tc.suffix, "", ordinary, http.StatusUnauthorized)
	}
	expectPublic()
	if got := adminList(); len(got) != 0 {
		t.Fatalf("unexpected initial state: %+v", got)
	}
	request(http.MethodPut, announcementAdmin+"/unknown", `{"text":"公告"}`, admin, http.StatusNotFound)
	request(http.MethodDelete, announcementAdmin+"/unknown", "", admin, http.StatusNotFound)
	request(http.MethodPost, announcementAdmin, `{"text":"   "}`, admin, http.StatusUnprocessableEntity)
	request(http.MethodPost, announcementAdmin, `{"text":"公告","href":"http://example.com/"}`, admin, http.StatusUnprocessableEntity)
	request(http.MethodPost, announcementAdmin, `{"text":`, admin, http.StatusBadRequest)

	first := adminMutation(http.MethodPost, "", `{"text":"  站点维护  ","href":" /welcome "}`)
	if !first.Active || first.ID == "" || first.Revision == "" || first.Text != "站点维护" || first.Href != "/welcome" {
		t.Fatalf("unexpected first publish: %+v", first)
	}
	second := adminMutation(http.MethodPost, "", `{"text":"第二条公告"}`)
	if !second.Active || second.ID == "" || second.ID == first.ID || second.Revision == first.Revision || second.Href != "" {
		t.Fatalf("unexpected second publish: %+v", second)
	}
	expectPublic(second, first)
	if got := adminList(); len(got) != 2 || got[0] != second || got[1] != first {
		t.Fatalf("unexpected initial list: %+v", got)
	}
	if got := adminMutation(http.MethodPut, first.ID, `{"text":"站点维护","href":"/welcome"}`); got != first {
		t.Fatalf("no-op save changed revision: %+v", got)
	}

	edited := adminMutation(http.MethodPut, first.ID, `{"text":"站点维护（时间更新）","href":"/welcome"}`)
	if !edited.Active || edited.ID != first.ID || edited.Revision == first.Revision || edited.Text != "站点维护（时间更新）" {
		t.Fatalf("text edit failed to update revision: %+v", edited)
	}
	expectPublic(second, edited)
	linked := adminMutation(http.MethodPut, first.ID, `{"text":"站点维护（时间更新）","href":"https://example.com/updates"}`)
	if !linked.Active || linked.ID != first.ID || linked.Revision == edited.Revision || linked.Text != edited.Text || linked.Href == edited.Href {
		t.Fatalf("link-only edit failed to update revision: %+v", linked)
	}
	expectPublic(second, linked)

	withdrawn := adminMutation(http.MethodDelete, first.ID, "")
	linked.Active = false
	if withdrawn != linked {
		t.Fatalf("withdrawal did not preserve content: got %+v, want %+v", withdrawn, linked)
	}
	if got := adminMutation(http.MethodDelete, first.ID, ""); got != withdrawn {
		t.Fatalf("repeated withdrawal changed record: %+v", got)
	}
	expectPublic(second)
	if got := adminList(); len(got) != 2 || got[0] != second || got[1] != withdrawn {
		t.Fatalf("admin list did not retain withdrawn entry: %+v", got)
	}

	republished := adminMutation(http.MethodPut, first.ID, `{"text":"站点维护（时间更新）","href":"https://example.com/updates"}`)
	if !republished.Active || republished.ID != first.ID || republished.Revision == withdrawn.Revision || republished.Text != withdrawn.Text || republished.Href != withdrawn.Href {
		t.Fatalf("republish did not create a revision: %+v", republished)
	}
	expectPublic(republished, second)

	// Each concurrent request has a distinct remote address so it exercises DB
	// serialization, not the unrelated shared-per-client rate limiter.
	t.Run("concurrent writes preserve other records", func(t *testing.T) {
		type write struct {
			method, id, body string
		}
		const count = 8
		run := func(writes []write) {
			t.Helper()
			type result struct {
				code int
				body string
			}
			done := make(chan result, len(writes))
			start := make(chan struct{})
			for i, w := range writes {
				go func() {
					<-start
					path := announcementAdmin
					if w.id != "" {
						path += "/" + w.id
					}
					req := httptest.NewRequest(w.method, path, strings.NewReader(w.body))
					req.RemoteAddr = fmt.Sprintf("192.0.2.%d:20000", i+1)
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+admin)
					resp := httptest.NewRecorder()
					router.ServeHTTP(resp, req)
					done <- result{resp.Code, resp.Body.String()}
				}()
			}
			close(start)
			for range writes {
				if resp := <-done; resp.code != http.StatusOK {
					t.Errorf("concurrent write failed (%d): %s", resp.code, resp.body)
				}
			}
		}
		creates := make([]write, count)
		for i := range creates {
			creates[i] = write{method: http.MethodPost, body: fmt.Sprintf(`{"text":"concurrent-%d"}`, i)}
		}
		run(creates)
		rows := adminList()
		if len(rows) != count+2 || rows[count] != republished || rows[count+1] != second {
			t.Fatalf("concurrent writes lost records or changed order: %+v", rows)
		}
		edits := make([]write, count)
		for i, row := range rows[:count] {
			edits[i] = write{method: http.MethodPut, id: row.ID, body: fmt.Sprintf(`{"text":"edited-%d"}`, i)}
		}
		run(edits)
		final := adminList()
		if len(final) != len(rows) {
			t.Fatalf("editing changed number of entries: %+v", final)
		}
		for i, row := range final[:count] {
			if row.ID != rows[i].ID || row.Revision == rows[i].Revision || row.Text != fmt.Sprintf("edited-%d", i) {
				t.Errorf("lost edit or changed creation order: %+v", row)
			}
		}
	})
}
