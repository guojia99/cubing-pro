package user

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrUserNotSoftDeleted = errors.New("用户尚未软删除，无法彻底删除")
	ErrUserInOrganizer    = errors.New("用户仍在主办团队中，请先移出")
)

// SoftDeleteUser 软删除用户（GORM Delete）。
func SoftDeleteUser(db *gorm.DB, userID uint) error {
	var u User
	if err := db.First(&u, userID).Error; err != nil {
		return err
	}
	if UserCubeStillInOrganizersTeam(db, u.CubeID) {
		return ErrUserInOrganizer
	}
	return db.Delete(&u).Error
}

// PurgeUser 彻底删除已软删除的用户及其附属数据。
func PurgeUser(db *gorm.DB, userID uint) error {
	var u User
	if err := db.Unscoped().First(&u, userID).Error; err != nil {
		return err
	}
	if !u.DeletedAt.Valid {
		return ErrUserNotSoftDeleted
	}
	if UserCubeStillInOrganizersTeam(db.Unscoped(), u.CubeID) {
		return ErrUserInOrganizer
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&CheckCode{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&UserKV{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Delete(&u).Error; err != nil {
			return fmt.Errorf("删除用户失败（可能仍有关联成绩/报名数据）: %w", err)
		}
		return nil
	})
}

// FindAdminUser 按 ID 或 CubeID 查找用户，可选包含软删除记录。
func FindAdminUser(db *gorm.DB, userID uint, cubeID string, unscoped bool) (User, error) {
	q := db
	if unscoped {
		q = q.Unscoped()
	}
	var u User
	switch {
	case userID != 0:
		err := q.First(&u, userID).Error
		return u, err
	case cubeID != "":
		err := q.Where("cube_id = ?", cubeID).First(&u).Error
		return u, err
	default:
		return u, fmt.Errorf("user_id 或 cube_id 必填")
	}
}
