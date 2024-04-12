package user

import (
	"net"
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

// User 用户信息表
type User struct {
	basemodel.Model

	// 账号信息
	Name            string `gorm:"unique;not null;column:name"` // 名称
	EnName          string `gorm:"column:en_name"`              // 英文名称
	LoginID         string `gorm:"column:login_id;unique;"`     // 登录账号
	CubeID          string `gorm:"column:cube_id;unique"`       // CubeID
	Password        string `gorm:"column:pw"`                   // 密码
	InitPassword    string `gorm:"column:init_pw"`              // 初始密码
	HistoryPassword string `gorm:"column:history_pw"`           // 历史密码
	Hash            string `gorm:"column:hash"`                 // 授权码

	// 状态信息
	Token              string     `gorm:"column:token"`                 // token
	LoginTime          time.Time  `gorm:"column:login_time"`            // 登录时间
	LoginIp            net.IP     `gorm:"column:login_ip"`              // 登录的IP
	Online             int        `gorm:"column:online"`                // 0 离线 1 在线 2 隐身
	ActivationTime     time.Time  `gorm:"column:a_time"`                // 启用时间
	BanReason          string     `gorm:"column:ban_reason"`            // 封禁原因
	SumPasswordWrong   int        `gorm:"column:sum_pw_wrong"`          // 累计尝试密码错误次数
	PassWordLockTime   *time.Time `gorm:"column:pw_lock_time"`          // 密码锁定时间
	LastUpdateNameTime *time.Time `gorm:"column:last_update_name_time"` // 上次修改名称时间

	// 媒体信息
	Sign       string `gorm:"column:sign"`        // 签名
	Avatar     string `gorm:"column:avatar"`      // 头像
	CoverPhoto string `gorm:"column:cover_photo"` // 封面相册

	// 等级信息
	Level         uint `gorm:"column:level"`   // 等级
	Experience    uint `gorm:"column:exp"`     // 经验
	UseExperience uint `gorm:"column:use_exp"` // 已消费经验值

	// 其他信息
	QQ           string `gorm:"column:qq"`            // qq号
	QQUniID      string `gorm:"column:qq_uni_id"`     // QQ唯一认证ID
	Wechat       string `gorm:"column:wechat"`        // 微信号
	WechatUnitID string `gorm:"column:wechat_uni_id"` // 微信唯一认证ID
	WcaID        string `gorm:"column:wca_id"`        // WCA ID
	Phone        string `gorm:"column:phone"`         // 手机号
	Email        string `gorm:"column:email"`         // 邮箱

	// 隐私信息
	ActualName  string    `gorm:"column:actual_name"` // 真实姓名
	Sex         int       `gorm:"column:sex"`         // 性别 0 无 1 男 2 女
	Nationality string    `gorm:"column:nationality"` // 国籍
	Province    string    `gorm:"column:province"`    // 省份、州
	Birthdate   time.Time `gorm:"column:birthdate"`   // 出生日期
	IDCard      string    `gorm:"column:id_card"`     // 身份证
	Address     string    `gorm:"column:address"`     // 地址

	// 代表信息
	DelegateName string `gorm:"column:represent_name"` // 代表称呼: 高级代表\代表\实习代表...
}
