package formulaforum_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSetDiscussionLifecycle(t *testing.T) {
	forum := newForumFixture(t, "333oh", "ZBLL", "T-ZBLL", "T-ZBLL 2GLL", "ADN11")
	request, post := forum.request, forum.post
	const sibling = "1234567890abcdef"
	const setPath = "/formula-forum/targets/set/" + set + "/comments"
	const siblingPath = "/formula-forum/targets/set/" + sibling + "/comments"
	const countsURL = "/public/formula-forum/counts?cube=333oh&class=ZBLL"
	if err := forum.db.Exec(`INSERT INTO formula_sets(id,class_id,name,position) VALUES(?, ?, 'U-ZBLL', 1)`, sibling, class).Error; err != nil {
		t.Fatal(err)
	}
	top := func(data map[string]any) map[string]any { return data["items"].([]any)[0].(map[string]any) }
	counts := func(path string, sets, groups, solutions map[string]any) {
		t.Helper()
		want := map[string]any{"sets": sets, "groups": groups, "solutions": solutions}
		if got := request("GET", path, "", "", 200); !reflect.DeepEqual(got, want) {
			t.Fatalf("counts: got %v, want %v", got, want)
		}
	}
	empty, groups, solutions := map[string]any{}, map[string]any{group: 1.0}, map[string]any{solution: 1.0}
	target := map[string]any{"kind": "set", "id": set, "cube": "333oh", "className": "ZBLL", "setId": set, "setName": "T-ZBLL", "groupName": "", "caseName": "", "moves": ""}
	counts(countsURL, empty, empty, empty)
	if thread := request("GET", "/public"+setPath, "", "", 200); thread["count"] != 0.0 || !reflect.DeepEqual(thread["target"], target) {
		t.Fatal(thread)
	}
	request("POST", "/user"+setPath, "", `{"body":"anonymous"}`, 401)
	for _, invalid := range []string{"T-ZBLL", strings.ToUpper(sibling), "z" + set[1:], set + "0"} {
		path := strings.Replace(setPath, set, invalid, 1)
		request("GET", "/public"+path, "", "", 400)
		request("POST", "/user"+path, forum.admin, `{"body":"bad id"}`, 400)
	}
	wrongID := strings.Replace(setPath, set, group, 1)
	request("GET", "/public"+wrongID, "", "", 404)
	request("POST", "/user"+wrongID, forum.admin, `{"body":"not a set"}`, 404)

	siblingRoot := post(siblingPath, forum.alice, "sibling set", 0)
	groupRoot := post(grp, forum.alice, "group", 0)
	post(sol, forum.alice, "solution", 0)
	root := post(setPath, forum.alice, "first set", 0)
	for _, wrongParent := range []int{siblingRoot, groupRoot} {
		request("POST", "/user"+setPath, forum.bob, fmt.Sprintf(`{"body":"wrong thread","parentId":%d}`, wrongParent), 404)
	}
	reply := post(setPath, forum.bob, "set reply", root)
	counts(countsURL, map[string]any{set: 2.0, sibling: 1.0}, groups, solutions)
	counts(strings.Replace(countsURL, "333oh", "333", 1), empty, empty, empty)
	counts(strings.Replace(countsURL, "ZBLL", "OLL", 1), empty, empty, empty)
	thread := request("GET", fmt.Sprintf("/public%s?size=1&commentId=%d", setPath, reply), "", "", 200)
	if thread["page"] != 2.0 || thread["count"] != 2.0 || top(thread)["id"] != float64(reply) || top(thread)["parentId"] != float64(root) || top(thread)["replyToName"] != "Alice" {
		t.Fatal(thread)
	}
	activity := request("GET", recent, "", "", 200)
	if activity["total"] != 4.0 || top(activity)["count"] != 2.0 || top(activity)["lastCommentId"] != float64(reply) || !reflect.DeepEqual(top(activity)["target"], target) {
		t.Fatal(activity)
	}
	request("GET", notices, "", "", 401)
	notice := request("GET", notices, forum.alice, "", 200)
	if notice["unread"] != 1.0 || top(notice)["commentId"] != float64(reply) || !reflect.DeepEqual(top(notice)["target"], target) {
		t.Fatal(notice)
	}
	read := fmt.Sprintf("%s/%.0f/read", notices, top(notice)["id"])
	request("PUT", read, forum.bob, "", 404)
	request("PUT", read, forum.alice, "", 200)
	forum.clientAddr = "192.0.2.2:1234"
	deleteRoot := fmt.Sprintf("%s/%d", mod, root)
	request("DELETE", deleteRoot, forum.bob, "", 401)
	request("DELETE", deleteRoot, forum.admin, "", 200)
	if thread = request("GET", "/public"+setPath, "", "", 200); thread["total"] != 2.0 || thread["count"] != 1.0 || top(thread)["deleted"] != true || top(thread)["body"] != "" {
		t.Fatal(thread)
	}
	request("DELETE", fmt.Sprintf("%s/%d", mod, reply), forum.admin, "", 200)
	counts(countsURL, map[string]any{sibling: 1.0}, groups, solutions)
	if request("GET", notices, forum.alice, "", 200)["total"] != 0.0 || request("GET", "/public"+setPath, "", "", 200)["total"] != 0.0 {
		t.Fatal("deleted reply or parent still visible")
	}

	siblingReply := post(siblingPath, forum.bob, "sibling reply", siblingRoot)
	if err := forum.db.Exec(`DELETE FROM formula_sets WHERE id = ?`, sibling).Error; err != nil {
		t.Fatal(err)
	}
	request("GET", "/public"+siblingPath, "", "", 404)
	request("POST", "/user"+siblingPath, forum.alice, `{"body":"removed"}`, 404)
	counts(countsURL, empty, groups, solutions)
	if got := request("GET", recent, "", "", 200); got["total"] != 2.0 {
		t.Fatal(got)
	}
	for _, view := range []map[string]any{request("GET", mod, forum.admin, "", 200), request("GET", notices, forum.alice, "", 200)} {
		missing := top(view)["target"].(map[string]any)
		if missing["kind"] != "set" || missing["id"] != sibling || missing["unavailable"] != true {
			t.Fatal(view)
		}
	}
	request("DELETE", fmt.Sprintf("%s/%d", mod, siblingReply), forum.admin, "", 200)
}
