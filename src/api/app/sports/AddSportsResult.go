package sports

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/sports"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CreateSportResultReq struct {
	EventId int     `json:"event_id"`
	UserId  int     `json:"user_id"`
	Result  float64 `json:"result"`
	Date    string  `json:"date"`
}

func CreateSportResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CreateSportResultReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var usr user.User
		if err := svc.DB.First(&usr, "id = ?", req.UserId).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}

		var event sports.SportEvent
		if err := svc.DB.First(&event, "id = ?", req.EventId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		var data = sports.SportResult{
			EventID:   event.ID,
			EventName: event.Name,
			UserID:    usr.ID,
			UserName:  usr.Name,
			CubeID:    usr.CubeID,
			Result:    req.Result,
			Date:      req.Date,
		}

		if err := svc.DB.Create(&data).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, data)
	}
}
