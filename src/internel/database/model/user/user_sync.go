package user

import (
	"fmt"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/post"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/sports"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

// PropagateUserName 将用户名同步到各表冗余字段。
func PropagateUserName(db *gorm.DB, userID uint, cubeID, name string) error {
	if name == "" {
		return nil
	}
	if cubeID != "" {
		if err := db.Model(&result.Results{}).Where("cube_id = ?", cubeID).Update("person_name", name).Error; err != nil {
			return err
		}
		if err := db.Model(&result.Record{}).Where("cube_id = ?", cubeID).Update("user_name", name).Error; err != nil {
			return err
		}
		if err := db.Model(&sports.SportResult{}).Where("cube_id = ?", cubeID).Update("user_name", name).Error; err != nil {
			return err
		}
	}
	if userID == 0 {
		return nil
	}
	if err := db.Model(&competition.Registration{}).Where("user_id = ?", userID).Update("user_name", name).Error; err != nil {
		return err
	}
	if err := db.Model(&post.Posts{}).Where("uid = ?", userID).Update("user_name", name).Error; err != nil {
		return err
	}
	if err := db.Model(&result.Results{}).Where("user_id = ?", userID).Update("person_name", name).Error; err != nil {
		return err
	}
	if err := db.Model(&result.Record{}).Where("user_id = ?", userID).Update("user_name", name).Error; err != nil {
		return err
	}
	return db.Model(&sports.SportResult{}).Where("user_id = ?", userID).Update("user_name", name).Error
}

// PropagateUserCubeID 将 CubeID 同步到各表冗余字段及主办团队引用。
func PropagateUserCubeID(db *gorm.DB, userID uint, oldCubeID, newCubeID, name string) error {
	if oldCubeID == "" || newCubeID == "" || oldCubeID == newCubeID {
		return nil
	}

	resultUpdates := map[string]interface{}{"cube_id": newCubeID}
	if name != "" {
		resultUpdates["person_name"] = name
	}
	if err := db.Model(&result.Results{}).Where("cube_id = ?", oldCubeID).Updates(resultUpdates).Error; err != nil {
		return err
	}

	nameUpdates := map[string]interface{}{"cube_id": newCubeID}
	if name != "" {
		nameUpdates["user_name"] = name
	}
	if err := db.Model(&result.Record{}).Where("cube_id = ?", oldCubeID).Updates(nameUpdates).Error; err != nil {
		return err
	}
	if err := db.Model(&sports.SportResult{}).Where("cube_id = ?", oldCubeID).Updates(nameUpdates).Error; err != nil {
		return err
	}

	if userID != 0 {
		if err := db.Model(&result.Results{}).Where("user_id = ?", userID).Update("cube_id", newCubeID).Error; err != nil {
			return err
		}
		if err := db.Model(&result.Record{}).Where("user_id = ?", userID).Update("cube_id", newCubeID).Error; err != nil {
			return err
		}
		if err := db.Model(&sports.SportResult{}).Where("user_id = ?", userID).Update("cube_id", newCubeID).Error; err != nil {
			return err
		}
	}

	return propagateOrganizerCubeID(db, oldCubeID, newCubeID)
}

func propagateOrganizerCubeID(db *gorm.DB, oldCubeID, newCubeID string) error {
	if err := db.Model(&Organizers{}).Where("leaderId = ?", oldCubeID).Update("leaderId", newCubeID).Error; err != nil {
		return err
	}

	var orgs []Organizers
	if err := db.Where("ass_org_users LIKE ?", fmt.Sprintf("%%%s%%", oldCubeID)).Find(&orgs).Error; err != nil {
		return err
	}
	for i := range orgs {
		var members []string
		if err := jsoniter.UnmarshalFromString(orgs[i].AssOrganizerUsers, &members); err != nil {
			continue
		}
		changed := false
		for j, member := range members {
			if strings.TrimSpace(member) == oldCubeID {
				members[j] = newCubeID
				changed = true
			}
		}
		if !changed {
			continue
		}
		if err := db.Model(&orgs[i]).Update("ass_org_users", utils.ToJSON(members)).Error; err != nil {
			return err
		}
	}
	return nil
}
