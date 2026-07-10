package users

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"gorm.io/gorm"
)

// todo 热门查询
func Users(svc *svc.Svc, maxSize int) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var out []user.User
		app_utils.GenerallyList(
			ctx, svc.DB, out, app_utils.ListSearchParam[user.User]{
				Model:   &user.User{},
				MaxSize: maxSize,
				CanSearchAndLike: []string{
					"cube_id", "en_name", "name",
				},
				Query:     "ban = ?",
				QueryCons: []interface{}{false},
				Select: []string{
					"id", "name", "en_name", "cube_id", "wca_id", "represent_name",
					"avatar",
				},
			},
		)
	}
}

func AdminUsers(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var out []user.User
		_, _ = app_utils.GenerallyList(
			ctx, svc.DB, out, app_utils.ListSearchParam[user.User]{
				Model:   &user.User{},
				MaxSize: 100,
				CanSearchAndLike: []string{
					"cube_id", "en_name", "name", "qq", "qq_uni_id", "id",
				},
				Select: nil,
			},
		)
	}
}

type CreateUserReq struct {
	Name string `json:"name"`

	// 非必填
	QQ         string `json:"qq"`
	ActualName string `json:"actualName"`
	WcaID      string `json:"wca_id"`
}

func CreateUser(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req CreateUserReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var newUser = user.User{
			Name:               req.Name,
			QQ:                 req.QQ,
			ActualName:         req.ActualName,
			WcaID:              req.WcaID,
			ActivationTime:     utils.PtrNow(),
			Hash:               string(utils.GenerateRandomKey(time.Now().UnixNano())),
			CubeID:             svc.Cov.GetCubeID(req.Name),
			InitPassword:       uuid.NewString(),
			LastUpdateNameTime: utils.PtrNow(),
		}
		newUser.SetAuth(user.AuthPlayer)

		if err := svc.DB.Create(&newUser).Error; err != nil {
			exception.ErrRegisterField.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

type UpdateUserNameReq struct {
	CubeID  string `json:"cube_id"`
	NewName string `json:"new_name"`
}

func UpdateUserName(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateUserNameReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		req.CubeID = strings.TrimSpace(req.CubeID)
		req.NewName = strings.TrimSpace(req.NewName)
		if req.CubeID == "" || req.NewName == "" {
			exception.ErrInvalidInput.ResponseWithError(ctx, "cube_id 与 new_name 不能为空")
			return
		}

		var curUser user.User
		if err := svc.DB.Where("cube_id = ?", req.CubeID).First(&curUser).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		var dup user.User
		if err := svc.DB.Unscoped().Where("name = ? AND id <> ?", req.NewName, curUser.ID).First(&dup).Error; err == nil {
			exception.ErrValidationFailed.ResponseWithError(ctx, "名称已被使用")
			return
		}

		curUser.Name = req.NewName
		curUser.LastUpdateNameTime = utils.PtrNow()

		err := svc.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&curUser).Error; err != nil {
				return err
			}
			return user.PropagateUserName(tx, curUser.ID, curUser.CubeID, curUser.Name)
		})
		if err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}

type UpdateUserCubeIDReq struct {
	CubeID    string `json:"cube_id"`
	NewCubeID string `json:"new_cube_id"`
}

func UpdateUserCubeID(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateUserCubeIDReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		req.CubeID = strings.TrimSpace(req.CubeID)
		req.NewCubeID = strings.TrimSpace(req.NewCubeID)
		if req.CubeID == "" || req.NewCubeID == "" {
			exception.ErrInvalidInput.ResponseWithError(ctx, "cube_id 与 new_cube_id 不能为空")
			return
		}
		if req.CubeID == req.NewCubeID {
			exception.ErrValidationFailed.ResponseWithError(ctx, "新 CubeID 与当前相同")
			return
		}
		if len(req.NewCubeID) != 10 {
			exception.ErrValidationFailed.ResponseWithError(ctx, "CubeID 长度须为 10")
			return
		}

		var curUser user.User
		if err := svc.DB.Where("cube_id = ?", req.CubeID).First(&curUser).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		var taken user.User
		if err := svc.DB.Unscoped().Where("cube_id = ?", req.NewCubeID).First(&taken).Error; err == nil {
			exception.ErrValidationFailed.ResponseWithError(ctx, "CubeID 已被占用")
			return
		}

		oldCubeID := curUser.CubeID
		curUser.CubeID = req.NewCubeID
		if strings.TrimSpace(curUser.LoginID) == oldCubeID {
			curUser.LoginID = req.NewCubeID
		}

		err := svc.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&curUser).Error; err != nil {
				return err
			}
			return user.PropagateUserCubeID(tx, curUser.ID, oldCubeID, req.NewCubeID, curUser.Name)
		})
		if err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}

type UpdateUserWCAIDReq struct {
	CubeID string `json:"cube_id"`
	WcaID  string `json:"wca_id"`
}

func UpdateUserWCAID(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateUserWCAIDReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var curUser user.User

		if err := svc.DB.Where("cube_id = ?", req.CubeID).First(&curUser).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		curUser.WcaID = strings.ToUpper(req.WcaID)
		if err := svc.DB.Save(&curUser).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

type AdminUserRefReq struct {
	UserID uint   `json:"user_id"`
	CubeID string `json:"cube_id"`
}

type UpdateUserQQReq struct {
	UserID  uint   `json:"user_id"`
	CubeID  string `json:"cube_id"`
	QQ      string `json:"qq"`
	QQUniID string `json:"qq_uni_id"`
}

func UpdateUserQQ(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateUserQQReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		req.CubeID = strings.TrimSpace(req.CubeID)
		req.QQ = strings.TrimSpace(req.QQ)
		req.QQUniID = strings.TrimSpace(req.QQUniID)

		curUser, err := user.FindAdminUser(svc.DB, req.UserID, req.CubeID, true)
		if err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		curUser.QQ = req.QQ
		curUser.QQUniID = req.QQUniID
		if err := svc.DB.Unscoped().Save(&curUser).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

func SoftDeleteUser(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AdminUserRefReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		req.CubeID = strings.TrimSpace(req.CubeID)

		curUser, err := user.FindAdminUser(svc.DB, req.UserID, req.CubeID, false)
		if err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		if err := user.SoftDeleteUser(svc.DB, curUser.ID); err != nil {
			if errors.Is(err, user.ErrUserInOrganizer) {
				exception.ErrValidationFailed.ResponseWithError(ctx, err.Error())
				return
			}
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

func PurgeUser(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AdminUserRefReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		req.CubeID = strings.TrimSpace(req.CubeID)

		curUser, err := user.FindAdminUser(svc.DB, req.UserID, req.CubeID, true)
		if err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		if err := user.PurgeUser(svc.DB, curUser.ID); err != nil {
			switch {
			case errors.Is(err, user.ErrUserNotSoftDeleted):
				exception.ErrValidationFailed.ResponseWithError(ctx, err.Error())
			case errors.Is(err, user.ErrUserInOrganizer):
				exception.ErrValidationFailed.ResponseWithError(ctx, err.Error())
			default:
				exception.ErrDatabase.ResponseWithError(ctx, err)
			}
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

type MergeUserReq struct {
	BaseUserCubeId   string `json:"base_user_cube_id"`
	MergedUserCubeId string `json:"merged_user_cube_id"`
}

func MergeUser(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req MergeUserReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		if err := svc.Cov.MergeUser(req.BaseUserCubeId, req.MergedUserCubeId); err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

type UpdateAuthReq struct {
	Set    bool      `json:"set"`
	CubeId string    `json:"cube_id"`
	Auth   user.Auth `json:"auth"`
}

func UpdateAuth(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateAuthReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		var curUser user.User
		if err := svc.DB.Where("cube_id = ?", req.CubeId).First(&curUser).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		if !req.Set && curUser.CheckAuth(user.AuthOrganizers) && req.Auth&user.AuthOrganizers != 0 {
			if user.UserCubeStillInOrganizersTeam(svc.DB, curUser.CubeID) {
				exception.ErrValidationFailed.ResponseWithError(ctx, "请先将该用户从主办团队的组长或成员中移出，再撤销主办权限")
				return
			}
		}

		if req.Set {
			curUser.SetAuth(req.Auth)
		} else {
			curUser.UnSetAuth(req.Auth)
		}
		if err := svc.DB.Save(&curUser).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
