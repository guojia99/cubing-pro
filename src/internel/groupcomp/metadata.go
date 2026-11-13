package groupcomp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"gorm.io/gorm"
)

type OrganizerOption struct {
	ID   uint   `json:"id"`
	Name string `json:"Name"`
}
type GroupOption struct {
	ID           uint   `json:"id"`
	Name         string `json:"Name"`
	OrganizersID uint   `json:"OrganizersID"`
}
type Filters struct {
	Organizers      []OrganizerOption `json:"organizers"`
	Groups          []GroupOption     `json:"groups"`
	Page            int               `json:"page"`
	Size            int               `json:"size"`
	OrganizersTotal int64             `json:"organizersTotal"`
	GroupsTotal     int64             `json:"groupsTotal"`
	OrganizersNext  bool              `json:"organizersNext"`
	GroupsNext      bool              `json:"groupsNext"`
}

func CompetitionNames(db *gorm.DB, orgID, groupID uint) (string, string, error) {
	var org user.Organizers
	var group competition.CompetitionGroup
	orgName, groupName := fmt.Sprintf("主办团队不存在（ID %d）", orgID), fmt.Sprintf("群组不存在（ID %d）", groupID)
	if err := db.Select("id", "name").First(&org, "id = ?", orgID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", err
	}
	if err := db.Select("id", "name").First(&group, "id = ?", groupID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", err
	}
	if org.ID != 0 {
		orgName = org.Name
		if strings.TrimSpace(orgName) == "" {
			orgName = fmt.Sprintf("未命名主办团队（ID %d）", orgID)
		}
	}
	if group.ID != 0 {
		groupName = group.Name
		if strings.TrimSpace(groupName) == "" {
			groupName = fmt.Sprintf("未命名群组（ID %d）", groupID)
		}
	}
	return orgName, groupName, nil
}

func FilterOptions(db *gorm.DB, page, size int) (Filters, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 200
	}
	if size > 500 {
		size = 500
	}
	out := Filters{Organizers: make([]OrganizerOption, 0), Groups: make([]GroupOption, 0), Page: page, Size: size}
	o := db.Model(&user.Organizers{})
	g := db.Model(&competition.CompetitionGroup{})
	if err := o.Count(&out.OrganizersTotal).Error; err != nil {
		return out, err
	}
	if err := g.Count(&out.GroupsTotal).Error; err != nil {
		return out, err
	}
	if err := o.Select("id", "name").Order("id ASC").Offset((page - 1) * size).Limit(size).Scan(&out.Organizers).Error; err != nil {
		return out, err
	}
	// Explicit SQL alias is needed for the existing mixed-case orgId column.
	if err := g.Select("id, name, orgId AS organizers_id").Order("id ASC").Offset((page - 1) * size).Limit(size).Scan(&out.Groups).Error; err != nil {
		return out, err
	}
	for i := range out.Organizers {
		if strings.TrimSpace(out.Organizers[i].Name) == "" {
			out.Organizers[i].Name = fmt.Sprintf("未命名主办团队（ID %d）", out.Organizers[i].ID)
		}
	}
	for i := range out.Groups {
		if strings.TrimSpace(out.Groups[i].Name) == "" {
			out.Groups[i].Name = fmt.Sprintf("未命名群组（ID %d）", out.Groups[i].ID)
		}
	}
	out.OrganizersNext = int64(page)*int64(size) < out.OrganizersTotal
	out.GroupsNext = int64(page)*int64(size) < out.GroupsTotal
	return out, nil
}
