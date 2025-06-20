package users

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
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
			fmt.Println(err)
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, kv)
	}
}

type SetKeyValueRequest struct {
	Type user.UserKVType

	Key   string
	Value string
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
