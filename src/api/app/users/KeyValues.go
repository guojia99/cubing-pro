package users

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetKeyValue(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		usr, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		key := ctx.Param("key")
		if key == "" {
			exception.ErrRequestBinding.ResponseWithError(ctx, "key is required")
			return
		}

		var kv user.UserKV
		err = svc.DB.Where("user_id = ? AND `key` = ?", usr.ID, key).First(&kv).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				exception.ErrResourceNotFound.ResponseWithError(ctx, err)
				return
			}
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, kv)
	}
}

// userKVListItem 列表接口不返回 value，仅元数据（含字节长度）
type userKVListItem struct {
	Key       string          `json:"key"`
	Type      user.UserKVType `json:"type"`
	ValueLen  int             `json:"valueLen" gorm:"column:value_len"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

func ListKeyValues(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		usr, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var rows []userKVListItem
		// LENGTH：按字节计，与 MaxKVLength 一致（MySQL / SQLite 均支持）
		err = svc.DB.Model(&user.UserKV{}).
			Select("`key`, `type`, `updated_at`, LENGTH(`value`) as value_len").
			Where("user_id = ?", usr.ID).
			Order("`key` ASC").
			Scan(&rows).Error
		if err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, rows)
	}
}

type SetKeyValueRequest struct {
	Type  user.UserKVType `json:"type"`
	Key   string          `json:"key"`
	Value string          `json:"value"`
}

func SetKeyValue(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		usr, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var req SetKeyValueRequest
		if err = ctx.ShouldBindJSON(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if !user.IsInWhitelist(req.Key) {
			exception.ErrRequestBinding.ResponseWithError(ctx, "this key is not in whitelist")
			return
		}

		if len(req.Value) > user.MaxKVLength {
			exception.ErrRequestBinding.ResponseWithError(ctx, "Max value length exceeded")
			return
		}

		kv := &user.UserKV{
			UserId: usr.ID,
			Key:    req.Key,
			Value:  req.Value,
			Type:   req.Type,
		}

		err = svc.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "key"}},   // 联合唯一约束字段
			DoUpdates: clause.AssignmentColumns([]string{"value", "type"}), // 更新字段
		}).Create(kv).Error

		if err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
