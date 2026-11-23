package formulaforum_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gin-gonic/gin"
	forumapi "github.com/guojia99/cubing-pro/src/api/app/formulaforum"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/api/routes"
	"github.com/guojia99/cubing-pro/src/internel/algs"
	"github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDiscussionLifecycle(t *testing.T) {
	t.Cleanup(func() {
		if err := os.Remove("wca_seniors_cache.json"); err != nil && !os.IsNotExist(err) {
			t.Error(err)
		}
	})
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "forum.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err = algs.InitCatalog(db, ""); err != nil {
		t.Fatal(err)
	}
	if err = formulaforum.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&user.User{}); err != nil {
		t.Fatal(err)
	}

	class, set := strings.Repeat("1", 16), strings.Repeat("2", 16)
	group, caseID, solution := strings.Repeat("3", 16), strings.Repeat("4", 16), strings.Repeat("a", 16)
	for _, row := range []struct {
		table  string
		values map[string]any
	}{
		{"formula_classes", map[string]any{"id": class, "cube": "333", "name": "OLL", "position": 0}},
		{"formula_sets", map[string]any{"id": set, "class_id": class, "name": "OLL", "position": 0}},
		{"formula_groups", map[string]any{"id": group, "set_id": set, "name": "T [2]", "position": 0}},
		{"formula_cases", map[string]any{"id": caseID, "group_id": group, "name": "OLL 45", "position": 0}},
		{"formula_solutions", map[string]any{"id": solution, "case_id": caseID, "position": 0, "moves": "F (R U R' U') F'"}},
	} {
		if err = db.Table(row.table).Create(row.values).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := &svc.Svc{DB: db}
	s.Cfg.GlobalConfig.Debug = true
	middleware.InitJWT(s)
	middleware.InitCheckAuth(s)
	router := gin.New()
	api := router.Group("")
	routes.AuthRouters(api, s)
	routes.PublicRouters(api, s)
	routes.UserRouters(api, s)
	routes.AdminRouters(api, s)

	clientAddr := "192.0.2.1:1234"
	request := func(method, path, token, body string, want int) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.RemoteAddr = clientAddr
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != want {
			t.Fatalf("%s %s: got %d, want %d; %s", method, path, resp.Code, want, resp.Body.String())
		}
		var data map[string]any
		if err := json.Unmarshal(resp.Body.Bytes(), &data); err != nil {
			t.Fatal(err)
		}
		if item, ok := data["data"].(map[string]any); ok {
			return item
		}
		return data
	}
	login := func(name string, role user.Auth) string {
		t.Helper()
		u := user.User{Name: name, CubeID: name, LoginID: name, Auth: role, Password: "test-pass"}
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
		return request("POST", "/auth/login", "", fmt.Sprintf(`{"loginID":%q,"password":"test-pass"}`, name), 200)["token"].(string)
	}
	alice := login("Alice", user.AuthPlayer)
	bob := login("Bob", user.AuthPlayer)
	admin := login("Mod", user.AuthAdmin|user.AuthPlayer)
	sol := "/formula-forum/targets/solution/" + solution + "/comments"
	grp := "/formula-forum/targets/group/" + group + "/comments"
	countsURL := "/public/formula-forum/counts?cube=333&class=OLL"
	notices := "/user/formula-forum/notifications"
	mod := "/admin/formula-forum/comments"
	recent := "/public/formula-forum/recent"
	post := func(target, token, body string, parent int) int {
		t.Helper()
		args := map[string]any{"body": body}
		if parent > 0 {
			args["parentId"] = parent
		}
		b, err := json.Marshal(args)
		if err != nil {
			t.Fatal(err)
		}
		return int(request("POST", "/user"+target, token, string(b), 200)["id"].(float64))
	}

	for _, invalid := range []string{strings.Repeat(solution, 4), strings.ToUpper(solution), "g" + solution[1:]} {
		request("POST", "/user/formula-forum/targets/solution/"+invalid+"/comments", alice, `{"body":"test"}`, 400)
	}
	initial := request("GET", "/public"+sol+"?page=2", "", "", 200)
	if initial["count"] != 0.0 || initial["page"] != 1.0 || initial["target"].(map[string]any)["caseName"] != "OLL 45" {
		t.Fatalf("initial solution: %#v", initial)
	}
	request("POST", "/user"+sol, "", `{"body":"test"}`, 401)
	request("GET", mod, bob, "", 401)
	first := post(sol, alice, "第一条🧩", 0)
	other := post(grp, alice, "分组", 0)
	request("POST", "/user"+sol, bob, fmt.Sprintf(`{"body":"wrong target","parentId":%d}`, other), 404)
	blocked := request("POST", "/user"+sol, bob, fmt.Sprintf(`{"body":"刷单返利","parentId":%d}`, first), 422)
	if blocked["code"] != "keyword_blocked" {
		t.Fatal(blocked)
	}
	request("POST", "/user"+sol, bob, fmt.Sprintf(`{"body":%q}`, strings.Repeat("字", 401)), 422)
	replyBody := strings.Repeat("妙", 399) + "🧩"
	reply := post(sol, bob, replyBody, first)
	self := post(sol, bob, "补充", reply)

	items := func(data map[string]any) []any { return data["items"].([]any) }
	top := func(data map[string]any) map[string]any { return items(data)[0].(map[string]any) }
	counts := request("GET", countsURL, "", "", 200)
	if counts["groups"].(map[string]any)[group] != 1.0 || counts["solutions"].(map[string]any)[solution] != 3.0 {
		t.Fatal(counts)
	}
	page := request("GET", fmt.Sprintf("/public%s?commentId=%d&size=1&page=99", sol, reply), "", "", 200)
	if page["page"] != 2.0 || top(page)["replyToName"] != "Alice" || top(page)["id"] != float64(reply) || top(page)["body"] != replyBody {
		t.Fatal(page)
	}
	page = request("GET", fmt.Sprintf("/public%s?commentId=%d", grp, reply), "", "", 200)
	if page["page"] != 1.0 || page["total"] != 1.0 || top(page)["id"] != float64(other) {
		t.Fatal(page)
	}
	notify := request("GET", notices, alice, "", 200)
	if notify["unread"] != 1.0 || top(notify)["commentId"] != float64(reply) || top(notify)["body"] != replyBody {
		t.Fatal(notify)
	}
	if request("GET", notices, bob, "", 200)["total"] != 0.0 {
		t.Fatal("self reply notified")
	}
	readURL := fmt.Sprintf("%s/%.0f/read", notices, top(notify)["id"])
	request("PUT", readURL, bob, "", 404)
	request("PUT", readURL, alice, "", 200)
	request("PUT", readURL, alice, "", 200)
	if request("GET", notices, alice, "", 200)["unread"] != 0.0 {
		t.Fatal("read notification counted as unread")
	}
	if view := request("GET", mod, admin, "", 200); top(view)["id"] != float64(self) || top(view)["target"].(map[string]any)["id"] != solution {
		t.Fatal(view)
	}

	del := func(id int) { t.Helper(); request("DELETE", fmt.Sprintf("%s/%d", mod, id), admin, "", 200) }
	// Removing the latest reply must clear its unread notification and expose
	// the previous visible activity, even when the discussion is still active.
	latestGroup := post(grp, bob, "另一组的留言", 0)
	unreadReply := post(sol, bob, "又试了一次", first)
	if view := request("GET", notices+"?page=2&size=1", alice, "", 200); view["page"] != 2.0 || view["unread"] != 1.0 || top(view)["commentId"] != float64(reply) {
		t.Fatal(view)
	}
	del(unreadReply)
	if view := request("GET", notices+"?page=2&size=1", alice, "", 200); view["page"] != 1.0 || view["unread"] != 0.0 || view["total"] != 1.0 || top(view)["commentId"] != float64(reply) {
		t.Fatal(view)
	}
	if view := request("GET", recent, "", "", 200); top(view)["target"].(map[string]any)["id"] != group || top(view)["lastCommentId"] != float64(latestGroup) {
		t.Fatal(view)
	}
	del(latestGroup)
	if view := request("GET", "/public"+grp+"?page=2&size=1", "", "", 200); view["page"] != 1.0 || view["total"] != 1.0 || top(view)["id"] != float64(other) {
		t.Fatal(view)
	}
	del(first)
	thread := request("GET", "/public"+sol, "", "", 200)
	if thread["count"] != 2.0 || thread["total"] != 3.0 || top(thread)["deleted"] != true || top(thread)["body"] != "" || top(thread)["authorName"] != "" {
		t.Fatal(thread)
	}
	request("POST", "/user"+sol, bob, fmt.Sprintf(`{"body":"已删","parentId":%d}`, first), 404)
	if view := request("GET", recent, "", "", 200); top(view)["target"].(map[string]any)["id"] != solution || top(view)["count"] != 2.0 || top(view)["lastCommentId"] != float64(self) {
		t.Fatal(view)
	}
	del(reply)
	request("DELETE", fmt.Sprintf("%s/%d", mod, reply), admin, "", 404)
	if request("GET", notices, alice, "", 200)["total"] != 0.0 {
		t.Fatal("deleted reply exposed in notification")
	}
	// Only the immediate parent of the surviving reply still occupies a page.
	page = request("GET", fmt.Sprintf("/public%s?commentId=%d&page=99&size=1", sol, first), "", "", 200)
	if page["page"] != 2.0 || page["total"] != 2.0 || top(page)["id"] != float64(self) {
		t.Fatal(page)
	}
	page = request("GET", fmt.Sprintf("/public%s?commentId=%d&size=1", sol, reply), "", "", 200)
	if page["total"] != 2.0 || page["count"] != 1.0 || page["page"] != 1.0 || top(page)["id"] != float64(reply) || top(page)["deleted"] != true {
		t.Fatal(page)
	}
	page = request("GET", fmt.Sprintf("/public%s?commentId=%d&size=1", sol, self), "", "", 200)
	if page["page"] != 2.0 || top(page)["id"] != float64(self) || top(page)["replyToName"] != "" {
		t.Fatal(page)
	}
	if counts = request("GET", countsURL, "", "", 200); counts["solutions"].(map[string]any)[solution] != page["count"] {
		t.Fatal(counts)
	}
	if view := request("GET", mod, admin, "", 200); view["total"] != 2.0 || len(items(view)) != 2 || top(view)["id"] != float64(self) {
		t.Fatal(view)
	}
	del(self)
	if thread = request("GET", "/public"+sol, "", "", 200); thread["total"] != 0.0 || thread["count"] != 0.0 {
		t.Fatal(thread)
	}
	if view := request("GET", recent+"?page=2&size=1", "", "", 200); view["page"] != 1.0 || view["total"] != 1.0 || top(view)["target"].(map[string]any)["id"] != group || top(view)["lastCommentId"] != float64(other) {
		t.Fatal(view)
	}
	if view := request("GET", mod+"?page=2&size=1", admin, "", 200); view["page"] != 1.0 || view["total"] != 1.0 || top(view)["id"] != float64(other) {
		t.Fatal(view)
	}

	var discussion formulaforum.FormulaDiscussion
	if err := db.Where("group_id = ?", group).Take(&discussion).Error; err != nil {
		t.Fatal(err)
	}
	var author user.User
	if err := db.Where("login_id = ?", "Bob").Take(&author).Error; err != nil {
		t.Fatal(err)
	}
	seed := make([]formulaforum.FormulaComment, 19)
	for i := range seed {
		seed[i] = formulaforum.FormulaComment{DiscussionID: discussion.ID, AuthorID: author.ID, Body: "跟帖"}
	}
	if err := db.Create(&seed).Error; err != nil {
		t.Fatal(err)
	}
	last := post(grp, bob, "跟帖", 0)
	focused := request("GET", fmt.Sprintf("/public%s?commentId=%d", grp, last), "", "", 200)
	if focused["page"] != 2.0 || len(items(focused)) != 1 || top(focused)["id"] != float64(last) {
		t.Fatal(focused)
	}
	if view := request("GET", "/public"+grp+"?page=99", "", "", 200); view["page"] != 2.0 || len(items(view)) != 1 || top(view)["id"] != float64(last) {
		t.Fatal(view)
	}

	clientAddr = "192.0.2.2:1234"
	root := post(sol, alice, "测试解法", 0)
	replyID := post(sol, bob, "测试回复", root)
	request("POST", "/user"+grp, bob, `{"body":"超出限额"}`, 429)
	checkUnavailable := func(wantCase string) string {
		t.Helper()
		request("GET", "/public"+sol, "", "", 404)
		request("POST", "/user"+sol, alice, `{"body":"失效解法"}`, 404)
		if view := request("GET", countsURL, "", "", 200); len(view["solutions"].(map[string]any)) != 0 {
			t.Fatal(view)
		}
		if view := request("GET", recent, "", "", 200); view["total"] != 1.0 || len(items(view)) != 1 || top(view)["target"].(map[string]any)["id"] != group {
			t.Fatal(view)
		}
		comments := request("GET", mod+"?size=100", admin, "", 200)
		if comments["total"] != 23.0 || len(items(comments)) != 23 || top(comments)["id"] != float64(replyID) {
			t.Fatal(comments)
		}
		notification := request("GET", notices, alice, "", 200)
		if notification["total"] != 1.0 || notification["unread"] != 1.0 || len(items(notification)) != 1 || top(notification)["commentId"] != float64(replyID) {
			t.Fatal(notification)
		}
		for _, view := range []map[string]any{comments, notification} {
			target := top(view)["target"].(map[string]any)
			if target["unavailable"] != true || target["kind"] != "solution" || target["id"] != solution || target["caseName"] != wantCase || target["moves"] != "" {
				t.Fatalf("target: %#v", target)
			}
		}
		return fmt.Sprintf("%s/%.0f/read", notices, top(notification)["id"])
	}
	if err = db.Table("formula_solutions").Where("id = ?", solution).Update("moves", "").Error; err != nil {
		t.Fatal(err)
	}
	checkUnavailable("OLL 45")
	if err = db.Exec("DELETE FROM formula_solutions WHERE id = ?", solution).Error; err != nil {
		t.Fatal(err)
	}
	request("PUT", checkUnavailable(""), alice, "", 200)
	if view := request("GET", notices, alice, "", 200); view["total"] != 1.0 || view["unread"] != 0.0 || top(view)["readAt"] == nil {
		t.Fatal(view)
	}
	del(replyID)
	if view := request("GET", mod, admin, "", 200); view["total"] != 22.0 || top(view)["id"] != float64(root) {
		t.Fatal(view)
	}
	if view := request("GET", notices, alice, "", 200); view["total"] != 0.0 || view["unread"] != 0.0 || len(items(view)) != 0 {
		t.Fatal(view)
	}
}

func TestCommentRateLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		router := gin.New()
		var account user.User
		account.ID = 1
		router.POST("/comments", func(ctx *gin.Context) {
			ctx.Set("authUserKey", account)
		}, forumapi.CommentRateLimit(10, time.Minute), func(ctx *gin.Context) {
			ctx.Status(201)
		})
		send := func() *httptest.ResponseRecorder {
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, httptest.NewRequest("POST", "/comments", nil))
			return resp
		}
		post := func(want int) {
			t.Helper()
			resp := send()
			if resp.Code != want || (want == 429 && !strings.Contains(resp.Body.String(), "发言过于频繁，请稍后再试")) {
				t.Fatalf("user %d: got %d %s, want %d", account.ID, resp.Code, resp.Body.String(), want)
			}
		}
		post(201)
		time.Sleep(30 * time.Second)
		for range 9 {
			post(201)
		}
		post(429)
		account.ID = 2
		post(201)
		account.ID = 1
		time.Sleep(30 * time.Second)
		post(201)
		post(429)
		time.Sleep(time.Minute)
		codes := make(chan int, 20)
		for range cap(codes) {
			go func() { codes <- send().Code }()
		}
		statuses := map[int]int{}
		for range cap(codes) {
			statuses[<-codes]++
		}
		if statuses[201] != 10 || statuses[429] != 10 {
			t.Fatalf("concurrent requests after expiry: %v", statuses)
		}
		post(429)
	})
}
