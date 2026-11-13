package organizers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type CompListReq struct {
	ID      uint   `form:"id"`
	Name    string `form:"name"`
	Status  string `form:"status"`
	OrgID   uint   `form:"orgId"`
	GroupID uint   `form:"groupId"`
	IsDone  *bool  `form:"isDone"`
	Date    string `form:"date"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}

type CompListItem struct {
	ID             uint              `json:"id"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	StrID          string            `json:"StrId"`
	Name           string            `json:"Name"`
	Status         string            `json:"Status"`
	RejectMsg      string            `json:"RejectMsg"`
	IsDone         bool              `json:"IsDone"`
	OrganizersID   uint              `json:"OrganizersID"`
	GroupID        uint              `json:"GroupId"`
	CompStartTime  time.Time         `json:"CompStartTime"`
	CompEndTime    time.Time         `json:"CompEndTime"`
	EventMin       string            `json:"EventMin"`
	Genre          competition.Genre `json:"Genre"`
	CanPreResult   bool              `json:"CanPreResult"`
	Count          int64             `json:"Count"`
	OrganizersName string            `json:"OrganizersName"`
	GroupName      string            `json:"GroupName"`
	PendingCount   int64             `json:"PendingCount"`
}

func listComps(svc *svc.Svc, admin bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompListReq
		if err := ctx.ShouldBindQuery(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if admin && req.Status == "" {
			req.Status = competition.Reviewing
		}
		if req.Page < 1 {
			req.Page = 1
		}
		if req.Size < 1 {
			req.Size = 20
		}
		if req.Size > 100 {
			req.Size = 100
		}
		q := svc.DB.WithContext(ctx.Request.Context()).Model(&competition.Competition{}).Where("group_id > 0 AND genre = ?", competition.OnlineInformal)
		if !admin {
			org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
			q = q.Where("orgId = ?", org.ID)
		}
		if req.ID > 0 {
			q = q.Where("id = ?", req.ID)
		}
		if req.Name != "" {
			q = q.Where("name LIKE ?", "%"+req.Name+"%")
		}
		if req.Status != "" && req.Status != "all" {
			switch req.Status {
			case competition.Reviewing, competition.Running, competition.Reject, competition.Temporary, competition.Ban:
			default:
				exception.ErrRequestBinding.ResponseWithError(ctx, "无效比赛状态")
				return
			}
			q = q.Where("status = ?", req.Status)
		}
		if req.OrgID > 0 {
			q = q.Where("orgId = ?", req.OrgID)
		}
		if req.GroupID > 0 {
			q = q.Where("group_id = ?", req.GroupID)
		}
		if req.IsDone != nil {
			q = q.Where("is_done = ?", *req.IsDone)
		}
		if req.Date != "" {
			d, err := time.ParseInLocation("2006-01-02", req.Date, competition.Beijing)
			if err != nil {
				exception.ErrRequestBinding.ResponseWithError(ctx, "date须为YYYY-MM-DD")
				return
			}
			q = q.Where("comp_start_time < ? AND comp_end_time >= ?", d.AddDate(0, 0, 1), d)
		}
		var total int64
		if err := q.Count(&total).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		var comps []competition.Competition
		// Explicit local projection excludes comp_json and scramble paths/data.
		if err := q.Select("id", "created_at", "updated_at", "str_id", "name", "status", "reject_msg", "is_done", "orgId", "group_id", "comp_start_time", "comp_end_time", "event_min", "genre", "can_pre_result", "count").Order("created_at DESC, id DESC").Offset((req.Page - 1) * req.Size).Limit(req.Size).Find(&comps).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		items := make([]CompListItem, 0, len(comps))
		for _, c := range comps {
			item := CompListItem{ID: c.ID, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, StrID: c.StrId, Name: c.Name, Status: c.Status, RejectMsg: c.RejectMsg, IsDone: c.IsDone, OrganizersID: c.OrganizersID, GroupID: c.GroupID, CompStartTime: c.CompStartTime, CompEndTime: c.CompEndTime, EventMin: c.EventMin, Genre: c.Genre, CanPreResult: c.CanPreResult, Count: c.Count}
			orgName, groupName, err := groupcomp.CompetitionNames(svc.DB.WithContext(ctx.Request.Context()), c.OrganizersID, c.GroupID)
			if err != nil {
				exception.ErrDatabase.ResponseWithError(ctx, err)
				return
			}
			item.OrganizersName, item.GroupName = orgName, groupName
			if err := svc.DB.Model(&result.PreResults{}).Where("comp_id = ? AND finish = ?", c.ID, false).Count(&item.PendingCount).Error; err != nil {
				exception.ErrDatabase.ResponseWithError(ctx, err)
				return
			}
			items = append(items, item)
		}
		exception.ResponseOK(ctx, app_utils.GenerallyListResp{Items: items, Total: total, Next: int64(req.Page)*int64(req.Size) < total})
	}
}

func OrgCompList(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// This existing endpoint also serves non-group competitions: preserve HEAD's list.
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		var list []competition.Competition
		app_utils.GenerallyList(ctx, svc.DB, list, app_utils.ListSearchParam[competition.Competition]{Model: &competition.Competition{}, MaxSize: 0, Query: "orgId = ?", QueryCons: []interface{}{org.ID}, Select: []string{"str_id", "name", "city", "event_min", "genre", "count", "free_p", "comp_start_time", "comp_end_time", "reg_start_time", "reg_end_time", "reg_cancel_dl_time", "reg_restart_time", "orgId", "wca_url", "min_count", "count", "status", "reject_msg", "is_done"}})
	}
}
func Comps(svc *svc.Svc) gin.HandlerFunc { return listComps(svc, true) }
