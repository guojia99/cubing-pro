package user

import (
	"fmt"
	"net"
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
)

type Auth = int

const (
	AuthPlayer     Auth = 1 << iota // 选手
	AuthOrganizers                  // 主办
	AuthDelegates                   // 代表
	AuthAdmin                       // 管理员
	AuthSuperAdmin                  // 超级管理员
)

// User 用户信息表
type User struct {
	basemodel.Model

	// Auth
	Auth Auth `gorm:"column:auth" json:"Auth,omitempty"` // 权限等级

	// 账号信息
	Name            string `gorm:"unique;not null;column:name" json:"Name,omitempty"` // 名称
	EnName          string `gorm:"column:en_name;null" json:"EnName,omitempty"`       // 英文名称
	LoginID         string `gorm:"column:login_id;null" json:"LoginID,omitempty"`     // 登录账号
	CubeID          string `gorm:"column:cube_id;null" json:"CubeID,omitempty"`       // CubeID
	Password        string `gorm:"column:pw;null" json:"-"`                           // 密码
	HistoryPassword string `gorm:"column:history_pw;null" json:"-"`                   // 历史密码
	Hash            string `gorm:"column:hash;null" json:"-"`                         // 授权码 todo 预留

	// v2用户
	InitPassword   string     `gorm:"column:init_pw;null" json:"InitPassword,omitempty"` // 初始密码 v2预留的坑
	ActivationTime *time.Time `gorm:"column:a_time" json:"ActivationTime,omitempty"`     // 启用时间

	// 状态信息
	Token              string     `gorm:"column:token;null" json:"-"`                        // token
	LoginTime          *time.Time `gorm:"column:login_time;null" json:"LoginTime,omitempty"` // 登录时间
	LoginIp            net.IP     `gorm:"column:login_ip;null" json:"-"`                     // 登录的IP
	Online             int        `gorm:"column:online;null" json:"Online,omitempty"`        // 0 离线 1 在线 2 隐身
	Ban                bool       `gorm:"column:ban;null" json:"Ban,omitempty"`              // 封禁
	BanReason          string     `gorm:"column:ban_reason;null" json:"BanReason,omitempty"` // 封禁原因
	SumPasswordWrong   int        `gorm:"column:sum_pw_wrong;null" json:"-"`                 // 累计尝试密码错误次数
	PassWordLockTime   *time.Time `gorm:"column:pw_lock_time;null" json:"-"`                 // 密码锁定时间
	LastUpdateNameTime *time.Time `gorm:"column:last_update_name_time;null" json:"-"`        // 上次修改名称时间

	// 媒体信息
	Sign       string `gorm:"column:sign;null" json:"Sign,omitempty"`              // 签名
	Avatar     string `gorm:"column:avatar;null" json:"Avatar,omitempty"`          // 头像
	CoverPhoto string `gorm:"column:cover_photo;null" json:"CoverPhoto,omitempty"` // 封面相册

	// 等级信息
	Level         uint `gorm:"column:level;null" json:"Level,omitempty"`           // 等级
	Experience    uint `gorm:"column:exp;null" json:"Experience,omitempty"`        // 经验
	UseExperience uint `gorm:"column:use_exp;null" json:"UseExperience,omitempty"` // 已消费经验值

	// 其他信息
	QQ           string `gorm:"column:qq;null" json:"QQ,omitempty"`                      // qq号
	QQUniID      string `gorm:"column:qq_uni_id;null" json:"QQUniID,omitempty"`          // QQ唯一认证ID
	Wechat       string `gorm:"column:wechat;null" json:"Wechat,omitempty"`              // 微信号
	WechatUnitID string `gorm:"column:wechat_uni_id;null" json:"WechatUnitID,omitempty"` // 微信唯一认证ID
	WcaID        string `gorm:"column:wca_id;null" json:"WcaID,omitempty"`               // WCA ID
	Phone        string `gorm:"column:phone;null" json:"Phone,omitempty"`                // 手机号
	Email        string `gorm:"column:email;null" json:"Email,omitempty"`                // 邮箱

	// 隐私信息
	ActualName  string     `gorm:"column:actual_name;null" json:"ActualName,omitempty"`  // 真实姓名
	Sex         int        `gorm:"column:sex;null" json:"Sex,omitempty"`                 // 性别 0 无 1 男 2 女
	Nationality string     `gorm:"column:nationality;null" json:"Nationality,omitempty"` // 国籍
	Province    string     `gorm:"column:province;null" json:"Province,omitempty"`       // 省份、州
	Birthdate   *time.Time `gorm:"column:birthdate;null" json:"Birthdate,omitempty"`     // 出生日期
	IDCard      string     `gorm:"column:id_card;null" json:"-"`                         // 身份证
	Address     string     `gorm:"column:address;null" json:"-"`                         // 地址

	// 代表信息
	DelegateName string `gorm:"column:represent_name;null" json:"DelegateName,omitempty"` // 代表称呼: 高级代表\代表\实习代表...
}

// AuthOmits 不允许用户得知的字段
func (u *User) AuthOmits() []string {
	return []string{
		"pw",
		"history_pw",
		"hash",
		"init_pw",
		"a_time",
		"token",
		"id_card",
	}
}

func (u *User) CheckPassword(password string) error {
	// 封禁时间
	if u.PassWordLockTime != nil {
		if time.Now().Sub(*u.PassWordLockTime) < 0 {
			return fmt.Errorf("用户尝试密码过多，禁止登录到 %+v", u.PassWordLockTime)
		}
		u.PassWordLockTime = nil
	}
	if len(u.Password) == 0 || len(password) == 0 {
		return fmt.Errorf("密码无效, `%s` - `%s`", u.Password, password)
	}

	if password == u.Password {
		u.SumPasswordWrong = 0
		u.PassWordLockTime = nil
		return nil
	}

	// 封禁次数, 每五次封禁一次
	u.SumPasswordWrong += 1
	if u.SumPasswordWrong%5 == 0 {
		t := time.Now().Add(time.Minute * time.Duration(u.SumPasswordWrong))
		u.PassWordLockTime = &t
		return fmt.Errorf("尝试次数过多，已封禁到 %+v", u.PassWordLockTime)
	}
	return fmt.Errorf("密码错误")
}

func (u *User) CheckAuth(auth Auth) bool { return u.Auth&auth != 0 }
func (u *User) SetAuth(auth ...Auth) {
	for _, a := range auth {
		u.Auth |= a
	}
}
func (u *User) UnSetAuth(auth ...Auth) {
	for _, a := range auth {
		u.Auth &= ^a
	}
}
