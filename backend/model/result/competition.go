package result

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	basemodel "github.com/guojia99/cubing-pro/model/base"
	"github.com/guojia99/cubing-pro/model/user"
)

type genre uint

const (
	WCA            genre = iota + 1 // WCA认证比赛
	Official                        // 线下正式比赛
	OnlineOfficial                  // 线上正式比赛
	Informal                        // 线下非正式比赛
	OnlineInformal                  // 线上非正式比赛
)

type Competition struct {
	basemodel.StringIDModel

	Genre genre `gorm:"column:genre,type:string,not null"`

	// 详情
	InformationJSON string                   `gorm:"column:info,type:string"` // 说明 []CompetitionInformation JSON
	Information     []CompetitionInformation `gorm:"-"`                       // I18N

	// 主办
	Sponsors       []user.User // 主办
	SponsorGroupID uint        `gorm:"column:sponsor_group_id"` // 主办团队

	// 时间相关
	CompStartTime                  time.Time `gorm:"column:comp_start_time"`    // 比赛开始时间
	CompEndTime                    time.Time `gorm:"column:comp_end_time"`      // 比赛结束时间
	RegistrationStartTime          time.Time `gorm:"column:reg_start_time"`     // 报名开始时间
	RegistrationEndTime            time.Time `gorm:"column:reg_end_time"`       // 报名结束时间
	RegistrationCancelDeadlineTime time.Time `gorm:"column:reg_cancel_dl_time"` // 退赛截止时间
	RegistrationRestartTime        time.Time `gorm:"column:reg_restart_time"`   // 报名重开时间

	// WCA相关
	WCAUrl     string      `gorm:"column:wca_url"` // WCA 认证地址
	Represents []user.User // 代表
	Count      uint        `gorm:"column:count"` // 人数
}

func (c *Competition) AfterFind(tx *gorm.DB) (err error) {
	return jsoniter.UnmarshalFromString(c.InformationJSON, &c.Information)
}

type PayType = uint

const (
	PayTypeTest   PayType = iota + 1 // 测试接口
	PayTypeWeChat                    // 微信
	PayTypeAliPay                    // 支付宝
	PayTypeCash                      // 现金
	PayTypeOther                     // 其他
)

type CompetitionRegistration struct {
	basemodel.Model

	CompID   uint   `gorm:"column:comp_id"` // 比赛ID
	CompName string `gorm:"column:comp_name"`
	UserID   uint   `gorm:"column:user_id"` // 选手ID
	UserName string `gorm:"column:user_name"`

	RegistrationTime time.Time  `gorm:"column:reg_time"`    // 报名时间
	AcceptationTime  *time.Time `gorm:"column:acc_time"`    // 通过时间
	RetireTime       *time.Time `gorm:"column:retire_time"` // 退赛时间

	Payments     []Payment `gorm:"-"`               // 报名费 + 追加的项目报名费
	PaymentsJSON string    `gorm:"column:payments"` // []Event JSON
}

type Payment struct {
	Events []Event `json:"events"` // 报名项目

	PayType PayType `json:"payType"` // 支付类型
	Remark  string  `json:"remark"`  // 备注

	// 支付相关
	CreateTime   time.Time `json:"createTime"`   // 创建时间
	OrderNumber  string    `json:"orderNumber"`  // 订单号
	BaseResult   float64   `json:"baseResult"`   // 基础报名费
	EventResults []float64 `json:"eventResults"` // 需要支付金额, 按每个项目来算
	ActualResult float64   `json:"actualResult"` // 实际支付金额, 按所有基础报名费 + 项目

	// 退费相关
	RefundTime         *time.Time `json:"refundTime"`         // 退费时间
	RefundRatio        float64    `json:"refundRatio"`        // 退费比例
	RefundOrderNumber  string     `json:"refundOrderNumber"`  // 退费订单号
	RefundResult       float64    `json:"refundResult"`       // 需要退费金额
	ActualRefundResult float64    `json:"actualRefundResult"` // 实际退费金额
}

type CompetitionDiscussion struct {
	basemodel.Model
}

/*
GT/TG GP/PG GX/XG

ET/TE EP/PE EX/XE
CT/TC CP/PC CX/XC

QT/TQ TX/XT
YT/TY DT/TD PT/TP
DX/XD PX/XP OT/TO





CR/RC  CZ/ZC  CL/LC
ER/RE  EZ/ZE  EL/LE
GR/RG  GL/LG  GZ/ZG
HR/RH  HZ/ZH  LZ/ZL LR/RL QL/LQ
*/
