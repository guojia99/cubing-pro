package compertion

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type Genre uint

const (
	WCA            Genre = iota + 1 // WCA认证比赛
	Official                        // 线下正式比赛
	OnlineOfficial                  // 线上正式比赛
	Informal                        // 线下非正式比赛
	OnlineInformal                  // 线上非正式比赛
)

type Competition struct {
	basemodel.Model // 这里的ID需要符合条件

	StrId string `gorm:"column:str_id,uniqueIndex"`

	// 详情
	Name         string             `gorm:"column:name"`          // 名称
	Illustrate   string             `gorm:"column:illustrate"`    // 详细说明
	Location     string             `gorm:"column:location"`      // 地址
	LocationAddr string             `gorm:"column:location_addr"` // 经纬坐标
	Country      string             `gorm:"column:country"`       // 地区
	City         string             `gorm:"column:city"`          // 城市
	RuleMD       string             `gorm:"column:rule_md"`       // 规则
	EventsJSON   string             `gorm:"column:events_json"`   // 项目列表JSON
	Events       []CompetitionEvent `gorm:"-"`                    // 项目列表
	Series       string             // 系列赛

	// 基础限制
	Genre           Genre `gorm:"column:genre,not null"` // 比赛形式
	MinCount        uint  `gorm:"column:min_count"`      // 最低开赛限制
	Count           uint  `gorm:"column:count"`          // 人数
	FreeParticipate bool  `gorm:"free_p"`                // 自由参赛, 支持非正式赛

	// 时间相关
	CompStartTime                  time.Time `gorm:"column:comp_start_time"`    // 比赛开始时间
	CompEndTime                    time.Time `gorm:"column:comp_end_time"`      // 比赛结束时间
	RegistrationStartTime          time.Time `gorm:"column:reg_start_time"`     // 报名开始时间
	RegistrationEndTime            time.Time `gorm:"column:reg_end_time"`       // 报名结束时间
	RegistrationCancelDeadlineTime time.Time `gorm:"column:reg_cancel_dl_time"` // 退赛截止时间
	RegistrationRestartTime        time.Time `gorm:"column:reg_restart_time"`   // 报名重开时间

	// 主办
	SponsorGroupID uint `gorm:"column:sponsor_group_id"` // 主办团队
	// WCA相关
	WCAUrl string `gorm:"column:wca_url"` // WCA 认证地址
}

type AssCompetitionUsers struct {
	basemodel.Model

	CompId       uint `gorm:"index:,unique,composite:AssCompetitionUsers"`
	SponsorsId   uint `gorm:"index:,unique,composite:AssCompetitionUsers"`
	RepresentsId uint `gorm:"index:,unique,composite:AssCompetitionUsers"`
}

func (c *Competition) AfterFind(tx *gorm.DB) (err error) {
	return jsoniter.UnmarshalFromString(c.EventsJSON, &c.Events)
}