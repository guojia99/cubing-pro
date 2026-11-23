package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/testutil"
)

const (
	announcementPublic = "/public/site-announcement"
	announcementAdmin  = "/admin/system_result/site-announcement"
)

type announcementResponse struct {
	ID          string    `json:"id"`
	Active      bool      `json:"active"`
	Revision    string    `json:"revision"`
	Text        string    `json:"text"`
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"publishedAt"`
}

func TestSiteAnnouncementRoutes(t *testing.T) {
	db := testutil.OpenMySQL(t)
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
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil || len(fields) != 6 || fields["active"] == nil || fields["href"] != nil {
			t.Fatalf("unexpected admin fields: %s (%v)", raw, err)
		}
		var out announcementResponse
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		if out.PublishedAt.IsZero() || out.PublishedAt.Location() != time.UTC {
			t.Fatalf("invalid UTC publication time: %s", raw)
		}
		return out
	}
	adminList := func() []announcementResponse {
		t.Helper()
		raw := request(http.MethodGet, announcementAdmin, "", admin, http.StatusOK)
		var fields []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		for _, entry := range fields {
			if len(entry) != 6 || entry["active"] == nil || entry["href"] != nil {
				t.Fatalf("unexpected admin list fields: %s", raw)
			}
		}
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
			if len(entry) != 5 || entry["id"] != want[i].ID || entry["revision"] != want[i].Revision || entry["text"] != want[i].Text || entry["title"] != want[i].Title || entry["publishedAt"] != want[i].PublishedAt.Format(time.RFC3339Nano) {
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
	request(http.MethodPut, announcementAdmin+"/unknown", `{"title":"标题","text":"公告"}`, admin, http.StatusNotFound)
	request(http.MethodDelete, announcementAdmin+"/unknown", "", admin, http.StatusNotFound)
	request(http.MethodPost, announcementAdmin, `{"text":`, admin, http.StatusBadRequest)

	createdAfter := time.Now().UTC()
	first := adminMutation(http.MethodPost, "", `{"title":"  维护通知  ","text":"  站点维护\n\n  第二行\r\n第三行  ","href":"http://ignored.example/"}`)
	if !first.Active || first.ID == "" || first.Revision == "" || first.Title != "维护通知" || first.Text != "站点维护\n\n  第二行\r\n第三行" || first.PublishedAt.Before(createdAfter) || first.PublishedAt.After(time.Now().UTC()) {
		t.Fatalf("unexpected first publish: %+v", first)
	}
	createdAfter = time.Now().UTC()
	second := adminMutation(http.MethodPost, "", `{"title":"第二条","text":"第二条公告"}`)
	if !second.Active || second.ID == "" || second.ID == first.ID || second.Revision == first.Revision || second.PublishedAt.Before(createdAfter) || second.PublishedAt.After(time.Now().UTC()) {
		t.Fatalf("unexpected second publish: %+v", second)
	}
	expectPublic(second, first)
	if got := adminList(); len(got) != 2 || got[0] != second || got[1] != first {
		t.Fatalf("unexpected initial list: %+v", got)
	}
	if got := adminMutation(http.MethodPut, first.ID, `{"title":"  维护通知 ","text":"\n站点维护\n\n  第二行\r\n第三行 \n"}`); got != first {
		t.Fatalf("normalized no-op changed record: %+v", got)
	}

	edited := adminMutation(http.MethodPut, first.ID, `{"title":"维护通知","text":"站点维护（时间更新）"}`)
	if !edited.Active || edited.ID != first.ID || edited.Revision == first.Revision || edited.Text != "站点维护（时间更新）" || edited.Title != first.Title || edited.PublishedAt != first.PublishedAt {
		t.Fatalf("body edit failed: %+v", edited)
	}
	expectPublic(second, edited)
	retitled := adminMutation(http.MethodPut, first.ID, `{"title":"维护时间更新","text":"站点维护（时间更新）"}`)
	if !retitled.Active || retitled.ID != first.ID || retitled.Revision == edited.Revision || retitled.Text != edited.Text || retitled.Title != "维护时间更新" || retitled.PublishedAt != first.PublishedAt {
		t.Fatalf("title-only edit failed: %+v", retitled)
	}
	expectPublic(second, retitled)

	withdrawn := adminMutation(http.MethodDelete, first.ID, "")
	retitled.Active = false
	if withdrawn != retitled {
		t.Fatalf("withdrawal did not preserve content: got %+v, want %+v", withdrawn, retitled)
	}
	if got := adminMutation(http.MethodDelete, first.ID, ""); got != withdrawn {
		t.Fatalf("repeated withdrawal changed record: %+v", got)
	}
	expectPublic(second)
	if got := adminList(); len(got) != 2 || got[0] != second || got[1] != withdrawn {
		t.Fatalf("admin list did not retain withdrawn entry: %+v", got)
	}

	// Separate the publication instants even on clocks with coarse resolution.
	time.Sleep(time.Millisecond)
	republishedAfter := time.Now().UTC()
	republished := adminMutation(http.MethodPut, first.ID, `{"title":"维护时间更新","text":"站点维护（时间更新）"}`)
	if !republished.Active || republished.ID != first.ID || republished.Revision == withdrawn.Revision || republished.Text != withdrawn.Text || republished.Title != withdrawn.Title || !republished.PublishedAt.After(withdrawn.PublishedAt) || republished.PublishedAt.Before(republishedAfter) || republished.PublishedAt.After(time.Now().UTC()) {
		t.Fatalf("republish did not refresh revision and publication time: %+v", republished)
	}
	expectPublic(republished, second)

	for _, tc := range []struct {
		name, title, text string
		valid             bool
	}{
		{"missing title", "", "正文", false},
		{"blank title", " \n\t", "正文", false},
		{"missing body", "标题", "", false},
		{"blank body", "标题", " \n\t", false},
		{"title over limit", strings.Repeat("🧩", 81), "正文", false},
		{"body over limit", "标题", strings.Repeat("🧩", 1501), false},
		{"minimum", " 魔 ", " 🧩 ", true},
		{"maximum", " " + strings.Repeat("🧩", 80) + " ", "\n" + strings.Repeat("魔", 1500) + "\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{"title": tc.title, "text": tc.text})
			if err != nil {
				t.Fatal(err)
			}
			if !tc.valid {
				request(http.MethodPost, announcementAdmin, string(body), admin, http.StatusUnprocessableEntity)
				request(http.MethodPut, announcementAdmin+"/"+second.ID, string(body), admin, http.StatusUnprocessableEntity)
				return
			}
			got := adminMutation(http.MethodPut, second.ID, string(body))
			if got.Title != strings.TrimSpace(tc.title) || got.Text != strings.TrimSpace(tc.text) || got.ID != second.ID || got.Revision == second.Revision || got.PublishedAt != second.PublishedAt {
				t.Fatalf("boundary content not saved correctly: %+v", got)
			}
			second = got
		})
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
			creates[i] = write{method: http.MethodPost, body: fmt.Sprintf(`{"title":"并发公告","text":"concurrent-%d"}`, i)}
		}
		run(creates)
		rows := adminList()
		if len(rows) != count+2 || rows[count] != republished || rows[count+1] != second {
			t.Fatalf("concurrent writes lost records or changed order: %+v", rows)
		}
		edits := make([]write, count)
		for i, row := range rows[:count] {
			edits[i] = write{method: http.MethodPut, id: row.ID, body: fmt.Sprintf(`{"title":"并发编辑","text":"edited-%d"}`, i)}
		}
		run(edits)
		final := adminList()
		if len(final) != len(rows) {
			t.Fatalf("editing changed number of entries: %+v", final)
		}
		if final[count] != republished || final[count+1] != second {
			t.Fatalf("concurrent edits changed independent records: %+v", final)
		}
		for i, row := range final[:count] {
			if row.ID != rows[i].ID || row.Revision == rows[i].Revision || row.Text != fmt.Sprintf("edited-%d", i) || row.Title != "并发编辑" || row.PublishedAt.IsZero() || row.PublishedAt != rows[i].PublishedAt {
				t.Errorf("lost edit or changed creation order: %+v", row)
			}
		}
	})
}
