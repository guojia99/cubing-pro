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
	Auth Auth `gorm:"column:auth"` // 权限等级

	// 账号信息
	Name            string `gorm:"unique;not null;column:name"` // 名称
	EnName          string `gorm:"column:en_name;null"`         // 英文名称
	LoginID         string `gorm:"column:login_id;null"`        // 登录账号
	CubeID          string `gorm:"column:cube_id;null"`         // CubeID
	Password        string `gorm:"column:pw;null"`              // 密码
	HistoryPassword string `gorm:"column:history_pw;null"`      // 历史密码
	Hash            string `gorm:"column:hash;null"`            // 授权码 todo 预留

	// v2用户
	InitPassword   string     `gorm:"column:init_pw;null"` // 初始密码 v2预留的坑
	ActivationTime *time.Time `gorm:"column:a_time"`       // 启用时间

	// 状态信息
	Token              string     `gorm:"column:token;null"`                 // token
	LoginTime          *time.Time `gorm:"column:login_time;null"`            // 登录时间
	LoginIp            net.IP     `gorm:"column:login_ip;null"`              // 登录的IP
	Online             int        `gorm:"column:online;null"`                // 0 离线 1 在线 2 隐身
	Ban                bool       `gorm:"column:ban;null"`                   // 封禁
	BanReason          string     `gorm:"column:ban_reason;null"`            // 封禁原因
	SumPasswordWrong   int        `gorm:"column:sum_pw_wrong;null"`          // 累计尝试密码错误次数
	PassWordLockTime   *time.Time `gorm:"column:pw_lock_time;null"`          // 密码锁定时间
	LastUpdateNameTime *time.Time `gorm:"column:last_update_name_time;null"` // 上次修改名称时间

	// 媒体信息
	Sign       string `gorm:"column:sign;null"`        // 签名
	Avatar     string `gorm:"column:avatar;null"`      // 头像
	CoverPhoto string `gorm:"column:cover_photo;null"` // 封面相册

	// 等级信息
	Level         uint `gorm:"column:level;null"`   // 等级
	Experience    uint `gorm:"column:exp;null"`     // 经验
	UseExperience uint `gorm:"column:use_exp;null"` // 已消费经验值

	// 其他信息
	QQ           string `gorm:"column:qq;null"`            // qq号
	QQUniID      string `gorm:"column:qq_uni_id;null"`     // QQ唯一认证ID
	Wechat       string `gorm:"column:wechat;null"`        // 微信号
	WechatUnitID string `gorm:"column:wechat_uni_id;null"` // 微信唯一认证ID
	WcaID        string `gorm:"column:wca_id;null"`        // WCA ID
	Phone        string `gorm:"column:phone;null"`         // 手机号
	Email        string `gorm:"column:email;null"`         // 邮箱

	// 隐私信息
	ActualName  string     `gorm:"column:actual_name;null"`      // 真实姓名
	Sex         int        `gorm:"column:sex;null"`              // 性别 0 无 1 男 2 女
	Nationality string     `gorm:"column:nationality;null"`      // 国籍
	Province    string     `gorm:"column:province;null"`         // 省份、州
	Birthdate   *time.Time `gorm:"column:birthdate;null"`        // 出生日期
	IDCard      string     `gorm:"column:id_card;null" json:"-"` // 身份证
	Address     string     `gorm:"column:address;null"`          // 地址

	// 代表信息
	DelegateName string `gorm:"column:represent_name;null"` // 代表称呼: 高级代表\代表\实习代表...
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
