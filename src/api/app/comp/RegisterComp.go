package comp

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	jsoniter "github.com/json-iterator/go"
)

type RegisterCompReq struct {
	CompReq

	Events []string `json:"Events"`
}

func RegisterComp(svc *svc.Svc) gin.HandlerFunc {
	// todo 这里暂时不考虑抢报名的情况， 反正也没人用。
	// todo 如果要考虑的话，就设置以比赛ID为单位的锁，对人数进行确认。
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		var req RegisterCompReq
		if err = app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		/*
			1. 查看比赛是否存在，并拉取最新的比赛列表。
				- todo 比赛级别的锁. 上锁
			2. 确认比赛是否还可以报名。
				- 时间相关.
					= 已经过了比赛报名时间的，或者没到的。
					= 比赛已经报满进入了等待报名重开时间的。
				- 项目不符合规范的
					= 如果参数里有本场不存在的项目直接报错。
					= 如果有资格线等筛选报名条件。
				- 确认该玩家是否已经有存在的报名, 如果已经有了报已报名
					= 需要主办审核的
					= 需要付费的。
					= 已经报名通过的。
				- 人数超过上限则无法报名。

				- todo 比赛级别的锁， 解锁
			3. 生成注册索引。
			4. 计算比赛付费金额， 查看是否需要付费.
				- 付费：生成付费链接。
			5. 保存数据库并返回。
		*/
		// todo 比赛可以用缓存
		var comp competition.Competition
		if err = svc.DB.First(&comp, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		// 时间相关的过滤
		if err = comp.CheckRegisterTime(); err != nil {
			exception.ErrCompNotRegister.ResponseWithError(ctx, err)
			return
		}
		// todo 比赛怎么重开后再次满了做限制？

		// 项目相关
		events := comp.EventMap()
		for _, event := range req.Events {
			// todo 资格线
			if _, ok := events[event]; !ok {
				exception.ErrCompNotRegister.ResponseWithError(ctx, fmt.Errorf("%s未在该比赛的项目列表", event))
				return
			}
		}
		allCost := comp.CompJSON.Cost.AllCost(time.Now(), req.Events)
		var reg competition.Registration
		// 如果已经存在订单，且等待支付， 则可能需要重新发起支付连接刷新
		if err = svc.DB.First(&reg, "comp_id = ? and user_id = ?", comp.ID, user.ID).Error; err == nil && reg.Status != competition.RegisterStatusWaitPayment {
			exception.ResponseOK(ctx, "已成功报名比赛，无需重新报名")
			return
		}

		// 获取当前人数
		// todo 这里要不要上锁？
		if comp.Count > 0 {
			var count int64
			svc.DB.Model(&competition.Registration{}).Where("comp_id = ?", comp.ID).Count(&count)
			if count >= comp.Count {
				exception.ErrCompNotRegister.ResponseWithError(ctx, "报名人数已达上限")
				comp.IsRegisterRestart = true
				svc.DB.Save(&comp)
				return
			}
		}

		// 初始化参数
		reg.CompID, reg.CompName, reg.UserID, reg.UserName = comp.ID, comp.Name, user.ID, user.Name
		if reg.RegistrationTime.IsZero() {
			reg.RegistrationTime = time.Now()
		}
		reg.Events, _ = jsoniter.MarshalToString(req.Events)
		reg.Status = competition.RegisterStatusWaitApply
		if allCost > 0 {
			reg.Status = competition.RegisterStatusWaitPayment
		} else if comp.AutomaticReview {
			reg.Status = competition.RegisterStatusPass
		}
		if allCost == 0 {
			if err = svc.DB.Save(&reg).Error; err != nil {
				exception.ErrCompNotRegister.ResponseWithError(ctx, err)
				return
			}
		}

		exception.ResponseOK(ctx, nil)
		// 付费相关
		// todo 暂不开发
	}
}
