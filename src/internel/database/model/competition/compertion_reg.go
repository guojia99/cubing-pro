package competition

import (
	"slices"
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	jsoniter "github.com/json-iterator/go"
)

type RegisterStatus = string

const (
	RegisterStatusPass        RegisterStatus = "pass"
	RegisterStatusQueue       RegisterStatus = "queue"
	RegisterStatusWaitPayment RegisterStatus = "wait_payment"
	RegisterStatusWaitApply   RegisterStatus = "wait_apply"
	RegisterStatusNotApply    RegisterStatus = "not_apply"
)

type Registration struct {
	basemodel.Model

	CompID   uint   `gorm:"column:comp_id"` // 比赛ID
	CompName string `gorm:"column:comp_name"`
	UserID   uint   `gorm:"column:user_id"` // 选手ID
	UserName string `gorm:"column:user_name"`

	Status RegisterStatus `gorm:"column:status"` // 注册状态

	RegistrationTime time.Time  `gorm:"column:reg_time"`    // 报名时间
	AcceptationTime  *time.Time `gorm:"column:acc_time"`    // 通过时间
	RetireTime       *time.Time `gorm:"column:retire_time"` // 退赛时间

	Events string `gorm:"column:events"` // 选择的项目ID列表

	Payments     []Payment `gorm:"-"`                        // 报名费 + 追加的项目报名费
	PaymentsJSON string    `gorm:"column:payments" json:"-"` // []Event JSON
}

func (c *Registration) SetEvent(event string) {

	list := c.EventsList()
	if slices.Contains(list, event) {
		return
	}

	list = append(list, event)
	c.Events, _ = jsoniter.MarshalToString(list)
	return
}

func (c *Registration) EventsList() []string {
	var out []string
	_ = jsoniter.UnmarshalFromString(c.Events, &out)
	return out
}

type PayType = uint

const (
	PayTypeTest   PayType = iota + 1 // 测试接口
	PayTypeWeChat                    // 微信
	PayTypeAliPay                    // 支付宝
	PayTypeCash                      // 现金
	PayTypeOther                     // 其他
)

type Payment struct {
	Events []event.Event `json:"events"` // 报名项目

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
