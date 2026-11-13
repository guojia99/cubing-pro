package groupcomp

import (
	"strings"
	"testing"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
)

func TestFiltersAndConsistentNames(t *testing.T) {
	db, _, _ := fixture(t)
	if err := db.AutoMigrate(&user.Organizers{}, &competition.CompetitionGroup{}); err != nil {
		t.Fatal(err)
	}
	for i := uint(1); i <= 3; i++ {
		org := user.Organizers{Model: basemodel.Model{ID: i}, Name: []string{"团队甲", "团队乙", "团队丙"}[i-1]}
		group := competition.CompetitionGroup{Model: basemodel.Model{ID: i}, Name: []string{"群组甲", "群组乙", "群组丙"}[i-1], OrganizersID: i}
		if err := db.Create(&org).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&group).Error; err != nil {
			t.Fatal(err)
		}
	}
	out, err := FilterOptions(db, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Organizers) != 2 || len(out.Groups) != 2 || out.OrganizersTotal != 3 || out.GroupsTotal != 3 || !out.GroupsNext || !out.OrganizersNext {
		t.Fatalf("wrong options page: %+v", out)
	}
	if out.Groups[0].OrganizersID != 1 || out.Groups[0].Name != "群组甲" {
		t.Fatal("group owner/name lost")
	}
	orgName, groupName, err := CompetitionNames(db, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if orgName != out.Organizers[0].Name || groupName != out.Groups[0].Name {
		t.Fatal("list/detail option names differ")
	}
	last, err := FilterOptions(db, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(last.Groups) != 1 || last.GroupsNext || last.OrganizersNext {
		t.Fatal("wrong final options page")
	}
	orgName, groupName, err = CompetitionNames(db, 99, 99)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(orgName, "不存在") || !strings.Contains(groupName, "不存在") {
		t.Fatal("missing association must have meaningful fallback")
	}
	capped, err := FilterOptions(db, 0, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if capped.Size != 500 || capped.Page != 1 {
		t.Fatal("unbounded option query")
	}
}
