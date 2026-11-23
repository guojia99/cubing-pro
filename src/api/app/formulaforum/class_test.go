package formulaforum_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestClassDiscussionLifecycle(t *testing.T) {
	forum := newForumFixture(t, "222", "EG", "EG1", "EG1-T", "T1")
	request, post := forum.request, forum.post
	const classPath = "/formula-forum/targets/class/" + class + "/comments"
	const countsPath = "/public/formula-forum/counts?cube=222&class=EG"
	counts := func(classCount, groupCount, solutionCount int) {
		t.Helper()
		groups, solutions := map[string]any{}, map[string]any{}
		if groupCount != 0 {
			groups[group] = float64(groupCount)
		}
		if solutionCount != 0 {
			solutions[solution] = float64(solutionCount)
		}
		got := request("GET", countsPath, "", "", 200)
		if got["classId"] != class || got["classCount"] != float64(classCount) || !reflect.DeepEqual(got["groups"], groups) || !reflect.DeepEqual(got["solutions"], solutions) {
			t.Fatal(got)
		}
	}
	target := map[string]any{"id": class, "kind": "class", "cube": "222", "className": "EG", "setName": "", "groupName": "", "caseName": "", "moves": ""}
	top := func(data map[string]any) map[string]any { return data["items"].([]any)[0].(map[string]any) }
	counts(0, 0, 0)
	if thread := request("GET", "/public"+classPath, "", "", 200); thread["count"] != 0.0 || !reflect.DeepEqual(thread["target"], target) {
		t.Fatal(thread)
	}
	request("POST", "/user"+classPath, "", `{"body":"anonymous"}`, 401)
	for _, invalid := range []string{"EG", strings.ToUpper(class), "z" + class[1:], class + "0"} {
		path := strings.Replace(classPath, class, invalid, 1)
		request("GET", "/public"+path, "", "", 400)
		request("POST", "/user"+path, forum.admin, `{"body":"bad id"}`, 400)
	}
	request("GET", "/public/formula-forum/targets/set/"+set+"/comments", "", "", 400)
	wrongID := strings.Replace(classPath, class, group, 1)
	request("GET", "/public"+wrongID, "", "", 404)
	request("POST", "/user"+wrongID, forum.admin, `{"body":"not a class"}`, 404)

	groupComment := post(grp, forum.alice, "group", 0)
	post(sol, forum.alice, "solution", 0)
	counts(0, 1, 1)
	request("POST", "/user"+classPath, forum.bob, fmt.Sprintf(`{"body":"wrong parent","parentId":%d}`, groupComment), 404)
	root := post(classPath, forum.alice, "overall EG", 0)
	request("POST", "/user"+grp, forum.bob, fmt.Sprintf(`{"body":"wrong parent","parentId":%d}`, root), 404)
	reply := post(classPath, forum.bob, "class reply", root)
	counts(2, 1, 1)
	thread := request("GET", fmt.Sprintf("/public%s?size=1&commentId=%d", classPath, reply), "", "", 200)
	if thread["page"] != 2.0 || thread["count"] != 2.0 || top(thread)["id"] != float64(reply) || top(thread)["parentId"] != float64(root) || top(thread)["replyToName"] != "Alice" {
		t.Fatal(thread)
	}
	activity := request("GET", recent, "", "", 200)
	if activity["total"] != 3.0 || top(activity)["count"] != 2.0 || top(activity)["lastCommentId"] != float64(reply) || !reflect.DeepEqual(top(activity)["target"], target) {
		t.Fatal(activity)
	}
	admin := request("GET", mod, forum.admin, "", 200)
	if top(admin)["id"] != float64(reply) || !reflect.DeepEqual(top(admin)["target"], target) {
		t.Fatal(admin)
	}
	request("GET", notices, "", "", 401)
	notice := request("GET", notices, forum.alice, "", 200)
	if notice["unread"] != 1.0 || top(notice)["commentId"] != float64(reply) || !reflect.DeepEqual(top(notice)["target"], target) {
		t.Fatal(notice)
	}
	read := fmt.Sprintf("%s/%.0f/read", notices, top(notice)["id"])
	request("PUT", read, forum.bob, "", 404)
	request("PUT", read, forum.alice, "", 200)
	deleteRoot := fmt.Sprintf("%s/%d", mod, root)
	request("DELETE", deleteRoot, forum.bob, "", 401)
	request("DELETE", deleteRoot, forum.admin, "", 200)
	counts(1, 1, 1)
	if thread = request("GET", "/public"+classPath, "", "", 200); thread["total"] != 2.0 || top(thread)["deleted"] != true || top(thread)["body"] != "" {
		t.Fatal(thread)
	}
	request("DELETE", fmt.Sprintf("%s/%d", mod, reply), forum.admin, "", 200)
	counts(0, 1, 1)
	if request("GET", notices, forum.alice, "", 200)["total"] != 0.0 || request("GET", "/public"+classPath, "", "", 200)["total"] != 0.0 {
		t.Fatal("deleted reply or parent still visible")
	}

	forum.clientAddr = "192.0.2.2:1234"
	// This second EG has the same label but no sets, groups or solutions.
	forum.clientAddr = "192.0.2.2:1234"
	const other = "1234567890abcdef"
	if err := forum.db.Exec(`INSERT INTO formula_classes(id,cube,name,position) VALUES(?, '333', 'EG', 0)`, other).Error; err != nil {
		t.Fatal(err)
	}
	otherPath := strings.Replace(classPath, class, other, 1)
	otherRoot := post(otherPath, forum.alice, "other cube", 0)
	otherReply := post(otherPath, forum.bob, "other reply", otherRoot)
	otherCounts := strings.Replace(countsPath, "222", "333", 1)
	if got := request("GET", otherCounts, "", "", 200); got["classId"] != other || got["classCount"] != 2.0 || len(got["groups"].(map[string]any)) != 0 || len(got["solutions"].(map[string]any)) != 0 {
		t.Fatal(got)
	}
	counts(0, 1, 1)
	if err := forum.db.Exec(`DELETE FROM formula_classes WHERE id = ?`, other).Error; err != nil {
		t.Fatal(err)
	}
	request("GET", "/public"+otherPath, "", "", 404)
	request("POST", "/user"+otherPath, forum.alice, `{"body":"removed"}`, 404)
	if request("GET", recent, "", "", 200)["total"] != 2.0 || request("GET", otherCounts, "", "", 200)["classId"] != "" {
		t.Fatal("missing catalog class included in public results")
	}
	for _, view := range []map[string]any{request("GET", mod, forum.admin, "", 200), request("GET", notices, forum.alice, "", 200)} {
		missing := top(view)["target"].(map[string]any)
		if missing["kind"] != "class" || missing["id"] != other || missing["unavailable"] != true {
			t.Fatal(view)
		}
	}
	request("DELETE", fmt.Sprintf("%s/%d", mod, otherReply), forum.admin, "", 200)
}
