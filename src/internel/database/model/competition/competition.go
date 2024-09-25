package competition

import (
	"errors"
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

type CompetitionStatus = string

const (
	Reviewing CompetitionStatus = "Reviewing"
	Running   CompetitionStatus = "Running"
	Reject    CompetitionStatus = "Reject"
	Temporary CompetitionStatus = "Temporary"
)

type Competition struct {
	basemodel.Model // 这里的ID需要符合条件

	StrId     string            `gorm:"column:str_id;null" json:"StrId,omitempty"`
	Status    CompetitionStatus `gorm:"column:status" json:"Status,omitempty"`
	RejectMsg string            `gorm:"column:reject_msg;null" json:"RejectMsg,omitempty"`

	// 详情
	Name         string          `gorm:"column:name" json:"Name,omitempty"`                       // 名称
	Illustrate   string          `gorm:"column:illustrate;null" json:"Illustrate,omitempty"`      // 详细说明 MD
	Location     string          `gorm:"column:location;null" json:"Location,omitempty"`          // 地址
	LocationAddr string          `gorm:"column:location_addr;null" json:"LocationAddr,omitempty"` // 经纬坐标
	Country      string          `gorm:"column:country;null" json:"Country,omitempty"`            // 地区
	City         string          `gorm:"column:city;null" json:"City,omitempty"`                  // 城市
	RuleMD       string          `gorm:"column:rule_md;null" json:"RuleMD,omitempty"`             // 规则
	CompJSONStr  string          `gorm:"column:comp_json;null" json:"-"`                          // 项目列表JSON
	CompJSON     CompetitionJson `gorm:"-" json:"comp_json,omitempty"`                            // 项目列表
	EventMin     string          `gorm:"column:event_min;null" json:"EventMin,omitempty"`         // 项目列表简列 ；隔开
	Series       string          `gorm:"column:series;null" json:"Series,omitempty"`              // 系列赛
	Logo         string          `gorm:"column:logo;null" json:"logo,omitempty"`                  // logo
	IsDone       bool            `gorm:"column:is_done"`                                          // 是否已经结束比赛

	// 基础限制
	Genre              Genre `gorm:"column:genre;not null" json:"Genre,omitempty"`                          // 比赛形式
	MinCount           int64 `gorm:"column:min_count;null" json:"MinCount,omitempty"`                       // 最低开赛限制
	Count              int64 `gorm:"column:count;null" json:"Count,omitempty"`                              // 最大人数
	FreeParticipate    bool  `gorm:"column:free_p;null" json:"FreeParticipate,omitempty"`                   // 自由参赛, 仅支持非正式赛
	AutomaticReview    bool  `gorm:"column:auto_review;null" json:"AutomaticReview,omitempty"`              // 自动审核
	CanPreResult       bool  `gorm:"column:can_pre_result;null" json:"CanPreResult,omitempty"`              // 允许提交预录入成绩
	CanStartedAddEvent bool  `gorm:"column:can_started_add_event;null" json:"CanStartedAddEvent,omitempty"` // 开赛后是否可追加项目（第一轮结束后不可追加）

	// 时间相关
	CompStartTime                  time.Time  `gorm:"column:comp_start_time" json:"CompStartTime,omitempty"`                          // 比赛开始时间
	CompEndTime                    time.Time  `gorm:"column:comp_end_time" json:"CompEndTime,omitempty"`                              // 比赛结束时间
	RegistrationStartTime          *time.Time `gorm:"column:reg_start_time;null" json:"RegistrationStartTime,omitempty"`              // 报名开始时间
	RegistrationEndTime            *time.Time `gorm:"column:reg_end_time;null" json:"RegistrationEndTime,omitempty"`                  // 报名结束时间
	RegistrationCancelDeadlineTime *time.Time `gorm:"column:reg_cancel_dl_time;null" json:"RegistrationCancelDeadlineTime,omitempty"` // 退赛截止时间
	IsRegisterRestart              bool       `gorm:"column:is_register_restart;null" json:"IsRegisterRestart,omitempty"`
	RegistrationRestartTime        *time.Time `gorm:"column:reg_restart_time;null" json:"RegistrationRestartTime,omitempty"` // 报名重开时间

	// 主办
	OrganizersID uint `gorm:"column:orgId;null" json:"OrganizersID,omitempty"` // 主办团队
	GroupID      uint `gorm:"column:group_id;null" json:"GroupId,omitempty"`   // 群ID

	// WCA相关
	WCAUrl string `gorm:"column:wca_url;null" json:"WCAUrl,omitempty"` // WCA 认证地址
}

type AssCompetitionSponsorsUsers struct {
	basemodel.Model

	CompId       uint `gorm:"index:,unique,composite:AssCompetitionSponsorsUsers"`
	SponsorsId   uint `gorm:"index:,unique,composite:AssCompetitionSponsorsUsers"`
	RepresentsId uint `gorm:"index:,unique,composite:AssCompetitionSponsorsUsers"`
}

func (c *Competition) AfterFind(tx *gorm.DB) (err error) {

	_ = jsoniter.UnmarshalFromString(c.CompJSONStr, &c.CompJSON)
	return nil
}

func (c *Competition) BeforeCreate(*gorm.DB) error { return c.update() }
func (c *Competition) BeforeUpdate(*gorm.DB) error { return c.update() }
func (c *Competition) BeforeSave(*gorm.DB) error   { return c.update() }
func (c *Competition) update() error {
	c.CompJSONStr, _ = jsoniter.MarshalToString(c.CompJSON)

	c.EventMin = ""
	for _, e := range c.CompJSON.Events {
		c.EventMin += e.EventName + ";"
	}

	return nil
}

func (c *Competition) EventMap() map[string]CompetitionEvent {
	var out = make(map[string]CompetitionEvent)
	for _, val := range c.CompJSON.Events {
		out[val.EventID] = val
	}
	return out
}

func (c *Competition) UpdateEvent(ev CompetitionEvent) {
	for n := range c.CompJSON.Events {
		if c.CompJSON.Events[n].EventID == ev.EventID {
			c.CompJSON.Events[n] = ev
			break
		}
	}
}

// IsRunningTime 是否在比赛时间段内
func (c *Competition) IsRunningTime() bool {
	if c.Status != Running {
		return false
	}
	if time.Since(c.CompStartTime) < 0 {
		return false
	}

	if time.Since(c.CompEndTime) > 0 {
		return false
	}
	return true
}

func (c *Competition) CheckRegisterTime() error {
	if c.RegistrationStartTime != nil && time.Since(*c.RegistrationRestartTime) < 0 {
		return errors.New("未到比赛报名开放时间")
	}
	if c.RegistrationEndTime != nil && time.Since(*c.RegistrationEndTime) > 0 {
		return errors.New("已过比赛注册报名时间")
	}
	if c.IsRegisterRestart && c.RegistrationRestartTime != nil && time.Since(*c.RegistrationRestartTime) < 0 {
		return errors.New("未到比赛重开报名时间")
	}
	return nil
}
