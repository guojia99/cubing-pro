package formulaforum_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/internel/convenient"
	"github.com/guojia99/cubing-pro/src/internel/database/model/formulaforum"
	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"gorm.io/gorm"
)

func TestPersonalFormulaLists(t *testing.T) {
	fixture := newForumFixture(t, "333", "OLL", "OLL", "T", "OLL 45")
	request, post, db := fixture.request, fixture.post, fixture.db
	comments, discussions := "/user/formula-forum/comments", "/user/formula-forum/discussions"
	request("GET", comments, "", "", 401)
	request("GET", discussions, "", "", 401)
	first := post(sol, fixture.alice, "我的评论", 0)
	reply := post(sol, fixture.bob, "回复", first)
	last := post(sol, fixture.alice, "我的回复", reply)
	other := post(grp, fixture.bob, "其他讨论", 0)
	items := func(data map[string]any) []any { return data["items"].([]any) }
	top := func(data map[string]any) map[string]any { return items(data)[0].(map[string]any) }

	view := request("GET", discussions, fixture.alice, "", 200)
	if view["total"] != 1.0 || top(view)["count"] != 3.0 || top(view)["lastCommentId"] != float64(last) {
		t.Fatal(view)
	}
	view = request("GET", comments+"?size=1&page=99", fixture.alice, "", 200)
	if view["total"] != 2.0 || view["page"] != 2.0 || top(view)["id"] != float64(first) {
		t.Fatal(view)
	}
	view = request("GET", comments, fixture.alice, "", 200)
	if top(view)["id"] != float64(last) || top(view)["parentId"] != float64(reply) || top(view)["replyToName"] != "Bob" {
		t.Fatal(view)
	}
	// Activity order uses timestamps rather than assuming IDs are chronological.
	newer := time.Now().Add(time.Hour)
	if err := db.Model(&formulaforum.FormulaComment{}).Where("id = ?", reply).Update("created_at", newer).Error; err != nil {
		t.Fatal(err)
	}
	view = request("GET", discussions, fixture.bob, "", 200)
	if view["total"] != 2.0 || top(view)["lastCommentId"] != float64(reply) || top(view)["target"].(map[string]any)["id"] != solution {
		t.Fatal(view)
	}
	if err := db.Table("formula_solutions").Where("id = ?", solution).Update("moves", "").Error; err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{comments, discussions} {
		view = request("GET", path, fixture.alice, "", 200)
		if top(view)["target"].(map[string]any)["unavailable"] != true {
			t.Fatal(view)
		}
	}
	for _, id := range []int{first, last} {
		request("DELETE", fmt.Sprintf("%s/%d", mod, id), fixture.admin, "", 200)
	}
	for _, path := range []string{comments, discussions} {
		view = request("GET", path+"?page=99", fixture.alice, "", 200)
		if view["total"] != 0.0 || view["page"] != 1.0 || len(items(view)) != 0 {
			t.Fatal(view)
		}
	}
	if view = request("GET", comments, fixture.bob, "", 200); view["total"] != 2.0 || top(view)["id"] != float64(reply) {
		t.Fatal(view, other)
	}
	// Repeating the normal migration must preserve data and the personal index.
	convenient.NewConvenient(db, false, configs.Config{})
	if !db.Migrator().HasIndex(&formulaforum.FormulaComment{}, "idx_formula_author_activity") {
		t.Fatal("missing personal comment index")
	}
}

func TestUnifiedNotifications(t *testing.T) {
	fixture := newForumFixture(t, "333", "OLL", "OLL", "T", "OLL 45")
	request, post, db := fixture.request, fixture.post, fixture.db
	inbox := "/user/notifications"
	request("GET", inbox, "", "", 401)
	request("GET", inbox+"?type=invalid", fixture.alice, "", 400)
	request("GET", inbox+"?size=101", fixture.alice, "", 400)
	first := post(sol, fixture.alice, "问题", 0)
	reply := post(sol, fixture.bob, "解答", first)
	post(sol, fixture.bob, "自己的回复", reply)
	base := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
	if err := db.Model(&formulaforum.FormulaComment{}).Where("id = ?", reply).Update("created_at", base.Add(time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	const announcementA = "11111111-1111-4111-8111-111111111111"
	const announcementB = "22222222-2222-4222-8222-222222222222"
	const announcementC = "33333333-3333-4333-8333-333333333333"
	announcements := []map[string]any{
		{"id": announcementA, "revision": "a-1", "title": "公告A", "text": "公告正文A", "active": true, "publishedAt": base.Add(2 * time.Minute)},
		{"id": announcementB, "revision": "b-1", "title": "公告B", "text": "公告正文B", "active": true, "publishedAt": base},
		{"id": announcementC, "revision": "c-1", "title": "公告C", "text": "撤回公告", "active": false, "publishedAt": base.Add(3 * time.Minute)},
	}
	writeAnnouncements(t, db, announcements)
	items := func(data map[string]any) []any { return data["items"].([]any) }
	top := func(data map[string]any) map[string]any { return items(data)[0].(map[string]any) }
	view := request("GET", inbox+"?size=1", fixture.alice, "", 200)
	if view["total"] != 3.0 || view["unread"] != 3.0 || top(view)["id"] != "site-announcement:"+announcementA || top(view)["body"] != "公告正文A" {
		t.Fatal(view)
	}
	legacy := request("GET", notices, fixture.alice, "", 200)
	legacyID := top(legacy)["id"].(float64)
	replyID := fmt.Sprintf("formula-reply:%.0f", legacyID)
	view = request("GET", inbox+"?size=1&page=2", fixture.alice, "", 200)
	if top(view)["id"] != replyID || top(view)["commentId"] != float64(reply) || top(view)["target"].(map[string]any)["id"] != solution {
		t.Fatal(view)
	}
	view = request("GET", inbox+"?type=site-announcement&size=1&page=99", fixture.alice, "", 200)
	if view["page"] != 2.0 || view["total"] != 2.0 || view["unread"] != 3.0 || top(view)["id"] != "site-announcement:"+announcementB {
		t.Fatal(view)
	}
	request("PUT", inbox+"/"+replyID+"/read", fixture.bob, "", 404)
	request("PUT", inbox+"/"+replyID+"/read", fixture.alice, "", 200)
	request("PUT", inbox+"/"+replyID+"/read", fixture.alice, "", 200)
	if view = request("GET", notices, fixture.alice, "", 200); view["unread"] != 0.0 || top(view)["readAt"] == nil {
		t.Fatal(view)
	}
	announcementRead := inbox + "/site-announcement:" + announcementA + "/read"
	request("PUT", announcementRead, fixture.alice, "", 200)
	request("PUT", announcementRead, fixture.alice, "", 200)
	view = request("GET", inbox, fixture.alice, "", 200)
	if view["unread"] != 1.0 || top(view)["readAt"] == nil {
		t.Fatal(view)
	}
	readAt := top(view)["readAt"]
	if view = request("GET", inbox, fixture.bob, "", 200); view["total"] != 2.0 || view["unread"] != 2.0 {
		t.Fatal(view)
	}
	announcements[0]["revision"], announcements[0]["text"] = "a-2", "修改后的公告"
	writeAnnouncements(t, db, announcements)
	view = request("GET", inbox, fixture.alice, "", 200)
	if view["unread"] != 1.0 || top(view)["readAt"] != readAt || top(view)["body"] != "修改后的公告" {
		t.Fatal(view)
	}
	announcements[0]["active"] = false
	writeAnnouncements(t, db, announcements)
	if view = request("GET", inbox, fixture.alice, "", 200); view["total"] != 2.0 || view["unread"] != 1.0 {
		t.Fatal(view)
	}
	request("PUT", announcementRead, fixture.alice, "", 404)
	announcements[0]["active"], announcements[0]["revision"], announcements[0]["publishedAt"] = true, "a-3", base.Add(4*time.Minute)
	writeAnnouncements(t, db, announcements)
	if view = request("GET", inbox, fixture.alice, "", 200); top(view)["readAt"] != readAt || view["unread"] != 1.0 {
		t.Fatal(view)
	}
	if err := db.Table("formula_solutions").Where("id = ?", solution).Update("moves", "").Error; err != nil {
		t.Fatal(err)
	}
	view = request("GET", inbox+"?type=formula-reply", fixture.alice, "", 200)
	if top(view)["target"].(map[string]any)["unavailable"] != true {
		t.Fatal(view)
	}
	request("DELETE", fmt.Sprintf("%s/%d", mod, reply), fixture.admin, "", 200)
	if view = request("GET", inbox+"?type=formula-reply", fixture.alice, "", 200); view["total"] != 0.0 || view["unread"] != 1.0 {
		t.Fatal(view)
	}
	for _, id := range []string{"invalid", "formula-reply:0", "formula-reply:01", "formula-reply:x", "site-announcement:"} {
		request("PUT", inbox+"/"+id+"/read", fixture.alice, "", 400)
	}
	request("PUT", inbox+"/site-announcement:invalid/read", fixture.alice, "", 404)
	request("PUT", inbox+"/site-announcement:44444444-4444-4444-8444-444444444444/read", fixture.alice, "", 404)
	var alice user.User
	if err := db.Where("name = ?", "Alice").Take(&alice).Error; err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Table("site_announcement_reads").Where("user_id = ? AND announcement_id = ?", alice.ID, announcementA).Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("read record count=%d err=%v", count, err)
	}
}

func writeAnnouncements(t *testing.T, db *gorm.DB, records []map[string]any) {
	t.Helper()
	if err := system.SetKeyJSONValue(db, "cubing-pro-site-announcements", records, "网站公告"); err != nil {
		t.Fatal(err)
	}
}
