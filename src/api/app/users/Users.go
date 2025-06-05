package users

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

// todo 热门查询
func Users(svc *svc.Svc, maxSize int) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var out []user.User
		app_utils.GenerallyList(
			ctx, svc.DB, out, app_utils.ListSearchParam{
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
			ctx, svc.DB, out, app_utils.ListSearchParam{
				Model:   &user.User{},
				MaxSize: 100,
				CanSearchAndLike: []string{
					"cube_id", "en_name", "name",
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

		var curUser user.User

		if err := svc.DB.Where("cube_id = ?", req.CubeID).First(&curUser).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		curUser.Name = req.NewName
		curUser.LastUpdateNameTime = utils.PtrNow()
		if err := svc.DB.Save(&curUser).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		svc.DB.Model(&result.Results{}).Where("cube_id = ?", curUser.CubeID).Update("person_name", curUser.Name)
		svc.DB.Model(&competition.Registration{}).Where("user_id = ?", curUser.ID).Update("user_name", curUser.Name)

		exception.ResponseOK(ctx, nil)
	}
}
