package organizers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"gorm.io/gorm"
)

func validOrganizersStatus(s user.OrganizersStatus) bool {
	switch s {
	case user.NotUse, user.Expired, user.Using, user.Applying, user.RejectApply,
		user.UnderAppeal, user.RejectAppeal, user.Disable, user.PermanentlyDisabled, user.Disband:
		return true
	}
	return false
}

func userExistsByCubeID(db *gorm.DB, cubeID string) bool {
	var n int64
	db.Model(&user.User{}).Where("cube_id = ?", cubeID).Count(&n)
	return n > 0
}

// --- 主办团队 CRUD ---

type adminCreateOrganizerBody struct {
	Name         string                `json:"name"`
	Introduction string                `json:"introduction"`
	Email        string                `json:"email"`
	LeaderCubeID string                `json:"leader_cube_id"`
	Status       user.OrganizersStatus `json:"status"`
	LeaderRemark string                `json:"leader_remark"`
	AdminMessage string                `json:"admin_message"`
	AssCubeIDs   []string              `json:"ass_cube_ids"`
}

func AdminCreateOrganizer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req adminCreateOrganizerBody
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			exception.ErrInvalidInput.ResponseWithError(ctx, "name 不能为空")
			return
		}
		if !validOrganizersStatus(req.Status) {
			exception.ErrInvalidInput.ResponseWithError(ctx, "status 无效")
			return
		}
		req.LeaderCubeID = strings.TrimSpace(req.LeaderCubeID)
		if req.LeaderCubeID != "" && !userExistsByCubeID(svc.DB, req.LeaderCubeID) {
			exception.ErrInvalidInput.ResponseWithError(ctx, "leader_cube_id 对应用户不存在")
			return
		}
		var dup int64
		svc.DB.Model(&user.Organizers{}).Where("name = ?", req.Name).Count(&dup)
		if dup > 0 {
			exception.ErrResultBeUse.ResponseWithError(ctx, "团队名称已存在")
			return
		}
		var ass []string
		for _, id := range req.AssCubeIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if !userExistsByCubeID(svc.DB, id) {
				exception.ErrInvalidInput.ResponseWithError(ctx, fmt.Sprintf("成员 cube_id 不存在: %s", id))
				return
			}
			if id == req.LeaderCubeID {
				continue
			}
			ass = append(ass, id)
		}
		ass = utils.RemoveDuplicates(ass)
		org := user.Organizers{
			Name:              req.Name,
			Introduction:      strings.TrimSpace(req.Introduction),
			Email:             strings.TrimSpace(req.Email),
			LeaderID:          req.LeaderCubeID,
			AssOrganizerUsers: utils.ToJSON(ass),
			Status:            req.Status,
			LeaderRemark:      strings.TrimSpace(req.LeaderRemark),
			AdminMessage:      strings.TrimSpace(req.AdminMessage),
		}
		if err := svc.DB.Create(&org).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, org)
	}
}

type adminOrgURI struct {
	OrgId uint `uri:"orgId"`
}

func AdminGetOrganizer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var uri adminOrgURI
		if err := app_utils.BindAll(ctx, &uri); err != nil {
			return
		}
		var org user.Organizers
		if err := svc.DB.First(&org, "id = ?", uri.OrgId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		var groups []competition.CompetitionGroup
		svc.DB.Where("orgId = ?", uri.OrgId).Find(&groups)
		exception.ResponseOK(ctx, gin.H{
			"organizer": org,
			"groups":    groups,
		})
	}
}

type adminUpdateOrganizerBody struct {
	adminOrgURI

	Name         *string                `json:"name"`
	Introduction *string                `json:"introduction"`
	Email        *string                `json:"email"`
	LeaderCubeID *string                `json:"leader_cube_id"`
	Status       *user.OrganizersStatus `json:"status"`
	LeaderRemark *string                `json:"leader_remark"`
	AdminMessage *string                `json:"admin_message"`
	AssCubeIDs   *[]string              `json:"ass_cube_ids"`
}

func AdminUpdateOrganizer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req adminUpdateOrganizerBody
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var org user.Organizers
		if err := svc.DB.First(&org, "id = ?", req.OrgId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if req.Name != nil {
			n := strings.TrimSpace(*req.Name)
			if n == "" {
				exception.ErrInvalidInput.ResponseWithError(ctx, "name 不能为空")
				return
			}
			var dup int64
			svc.DB.Model(&user.Organizers{}).Where("name = ? AND id <> ?", n, org.ID).Count(&dup)
			if dup > 0 {
				exception.ErrResultBeUse.ResponseWithError(ctx, "团队名称已存在")
				return
			}
			org.Name = n
		}
		if req.Introduction != nil {
			org.Introduction = strings.TrimSpace(*req.Introduction)
		}
		if req.Email != nil {
			org.Email = strings.TrimSpace(*req.Email)
		}
		if req.LeaderCubeID != nil {
			l := strings.TrimSpace(*req.LeaderCubeID)
			if l != "" && !userExistsByCubeID(svc.DB, l) {
				exception.ErrInvalidInput.ResponseWithError(ctx, "leader_cube_id 对应用户不存在")
				return
			}
			org.LeaderID = l
		}
		if req.Status != nil {
			if !validOrganizersStatus(*req.Status) {
				exception.ErrInvalidInput.ResponseWithError(ctx, "status 无效")
				return
			}
			org.Status = *req.Status
		}
		if req.LeaderRemark != nil {
			org.LeaderRemark = strings.TrimSpace(*req.LeaderRemark)
		}
		if req.AdminMessage != nil {
			org.AdminMessage = strings.TrimSpace(*req.AdminMessage)
		}
		if req.AssCubeIDs != nil {
			var ass []string
			for _, id := range *req.AssCubeIDs {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				if !userExistsByCubeID(svc.DB, id) {
					exception.ErrInvalidInput.ResponseWithError(ctx, fmt.Sprintf("成员 cube_id 不存在: %s", id))
					return
				}
				if org.LeaderID != "" && id == org.LeaderID {
					continue
				}
				ass = append(ass, id)
			}
			ass = utils.RemoveDuplicates(ass)
			org.AssOrganizerUsers = utils.ToJSON(ass)
		}
		if err := svc.DB.Save(&org).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, org)
	}
}

func AdminDeleteOrganizer(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var uri adminOrgURI
		if err := app_utils.BindAll(ctx, &uri); err != nil {
			return
		}
		var n int64
		svc.DB.Model(&competition.Competition{}).Where("orgId = ?", uri.OrgId).Count(&n)
		if n > 0 {
			exception.ErrValidationFailed.ResponseWithError(ctx, "该主办团队下仍有比赛记录，无法删除")
			return
		}
		if err := svc.DB.Delete(&user.Organizers{}, "id = ?", uri.OrgId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		svc.DB.Where("orgId = ?", uri.OrgId).Delete(&competition.CompetitionGroup{})
		exception.ResponseOK(ctx, nil)
	}
}

// --- 比赛群组 ---

type adminCreateGroupBody struct {
	adminOrgURI

	Name         string   `json:"name"`
	QQGroups     []string `json:"qq_groups"`
	QQGroupUid   []string `json:"qq_group_uid"`
	WechatGroups []string `json:"wechat_groups"`
}

func AdminListOrganizerGroups(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var uri adminOrgURI
		if err := app_utils.BindAll(ctx, &uri); err != nil {
			return
		}
		var list []competition.CompetitionGroup
		app_utils.GenerallyList(
			ctx, svc.DB, list, app_utils.ListSearchParam[competition.CompetitionGroup]{
				Model:     &competition.CompetitionGroup{},
				MaxSize:   20,
				Query:     "orgId = ?",
				QueryCons: []interface{}{uri.OrgId},
			},
		)
	}
}

func AdminCreateCompetitionGroup(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req adminCreateGroupBody
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var org user.Organizers
		if err := svc.DB.First(&org, "id = ?", req.OrgId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		var count int64
		svc.DB.Model(&competition.CompetitionGroup{}).Where("orgId = ?", req.OrgId).Count(&count)
		if int(count) >= user.MaxCompetitionGroupsPerOrganizer {
			exception.ErrValidationFailed.ResponseWithError(ctx, fmt.Sprintf("每个主办团队最多 %d 个比赛群组", user.MaxCompetitionGroupsPerOrganizer))
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			exception.ErrInvalidInput.ResponseWithError(ctx, "name 不能为空")
			return
		}
		g := competition.CompetitionGroup{
			Name:         req.Name,
			OrganizersID: req.OrgId,
			QQGroups:     competition.StringListToDB(req.QQGroups),
			QQGroupUid:   competition.StringListToDB(req.QQGroupUid),
			WechatGroups: competition.StringListToDB(req.WechatGroups),
		}
		if err := svc.DB.Create(&g).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, g)
	}
}

type adminGroupURI struct {
	GroupId uint `uri:"groupId"`
}

type adminUpdateGroupBody struct {
	adminGroupURI

	Name         *string   `json:"name"`
	QQGroups     *[]string `json:"qq_groups"`
	QQGroupUid   *[]string `json:"qq_group_uid"`
	WechatGroups *[]string `json:"wechat_groups"`
}

func AdminUpdateCompetitionGroup(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req adminUpdateGroupBody
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		var g competition.CompetitionGroup
		if err := svc.DB.First(&g, "id = ?", req.GroupId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if req.Name != nil {
			n := strings.TrimSpace(*req.Name)
			if n == "" {
				exception.ErrInvalidInput.ResponseWithError(ctx, "name 不能为空")
				return
			}
			g.Name = n
		}
		if req.QQGroups != nil {
			g.QQGroups = competition.StringListToDB(*req.QQGroups)
		}
		if req.QQGroupUid != nil {
			g.QQGroupUid = competition.StringListToDB(*req.QQGroupUid)
		}
		if req.WechatGroups != nil {
			g.WechatGroups = competition.StringListToDB(*req.WechatGroups)
		}
		if err := svc.DB.Save(&g).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, g)
	}
}

func AdminDeleteCompetitionGroup(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var uri adminGroupURI
		if err := app_utils.BindAll(ctx, &uri); err != nil {
			return
		}
		if err := svc.DB.Delete(&competition.CompetitionGroup{}, "id = ?", uri.GroupId).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

// --- 主办成员（主办管理员） ---

type adminMemberBody struct {
	adminOrgURI

	CubeID    string `json:"cube_id"`
	GrantAuth bool   `json:"grant_auth"`
}

func AdminAddOrganizerMember(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req adminMemberBody
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		req.CubeID = strings.TrimSpace(req.CubeID)
		if req.CubeID == "" {
			exception.ErrInvalidInput.ResponseWithError(ctx, "cube_id 不能为空")
			return
		}
		var org user.Organizers
		if err := svc.DB.First(&org, "id = ?", req.OrgId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		var u user.User
		if err := svc.DB.Where("cube_id = ?", req.CubeID).First(&u).Error; err != nil {
			exception.ErrUserNotFound.ResponseWithError(ctx, err)
			return
		}
		if org.LeaderID == req.CubeID {
			exception.ErrInvalidInput.ResponseWithError(ctx, "该用户已是主办组长，无需重复添加为成员")
			return
		}
		org.SetUsersCubingID([]string{req.CubeID})
		if err := svc.DB.Save(&org).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		if req.GrantAuth {
			u.SetAuth(user.AuthOrganizers)
			_ = svc.DB.Save(&u)
		}
		exception.ResponseOK(ctx, org)
	}
}

func AdminRemoveOrganizerMember(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var uri adminOrgURI
		if err := app_utils.BindAll(ctx, &uri); err != nil {
			return
		}
		cubeID := strings.TrimSpace(ctx.Query("cube_id"))
		if cubeID == "" {
			exception.ErrInvalidInput.ResponseWithError(ctx, "query cube_id 必填")
			return
		}
		var org user.Organizers
		if err := svc.DB.First(&org, "id = ?", uri.OrgId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if org.LeaderID == cubeID {
			exception.ErrValidationFailed.ResponseWithError(ctx, "不能移出主办组长，请先变更 leader_cube_id 或将组长设为空")
			return
		}
		org.DeleteUserID([]string{cubeID})
		if err := svc.DB.Save(&org).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, org)
	}
}
