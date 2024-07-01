package main

import "time"

// Score 成绩表
type Score struct {
	ID         uint      `json:"ID" gorm:"primaryKey;column:id"`
	CreatedAt  time.Time `json:"-" gorm:"autoCreateTime;column:created_at"`
	PlayerID   uint      `json:"PlayerID,omitempty" gorm:"index;not null;column:player_id"`   // 选手的ID
	PlayerName string    `json:"PlayerName,omitempty" gorm:"column:player_name"`              // 玩家名
	ContestID  uint      `json:"ContestID,omitempty" gorm:"index;not null;column:contest_id"` // 比赛的ID
	RouteID    uint      `json:"RouteID,omitempty" gorm:"index;column:route_id"`              // 轮次
	Project    string    `json:"Project,omitempty" gorm:"not null;column:project"`            // 分项目 333/222/444等
	Result1    float64   `json:"R1,omitempty" gorm:"column:r1;NULL"`                          // 成绩1 多盲时这个成绩是实际还原数
	Result2    float64   `json:"R2,omitempty" gorm:"column:r2;NULL"`                          // 成绩2 多盲时这个成绩是尝试复原数
	Result3    float64   `json:"R3,omitempty" gorm:"column:r3;NULL"`                          // 成绩3 多盲时这个成绩是计时
	Result4    float64   `json:"R4,omitempty" gorm:"column:r4;NULL"`                          // 成绩4
	Result5    float64   `json:"R5,omitempty" gorm:"column:r5;NULL"`                          // 成绩5
	Best       float64   `json:"Best,omitempty" gorm:"column:best;NULL"`                      // 五把最好成绩
	Avg        float64   `json:"Avg,omitempty" gorm:"column:avg;NULL"`                        // 五把平均成绩
}

// Contest 比赛表，记录某场比赛
type Contest struct {
	ID          uint      `json:"ID" gorm:"primaryKey;column:id"`
	CreatedAt   time.Time `json:"-" gorm:"autoCreateTime;column:created_at"`
	Name        string    `json:"Name,omitempty" gorm:"unique;not null;column:name"`        // 比赛名
	Type        string    `json:"Type,omitempty" gorm:"column:c_type"`                      // 类型 正式 | 线上 | 线下
	GroupID     string    `json:"GroupID,omitempty" gorm:"column:group_id"`                 // 线上群赛ID
	Description string    `json:"Description,omitempty" gorm:"not null;column:description"` // 描述
	IsEnd       bool      `json:"IsEnd,omitempty" gorm:"null;column:is_end"`                // 是否已结束
	RoundIds    string    `json:"RoundIds,omitempty" gorm:"column:round_ids"`               // 轮次ID
	StartTime   time.Time `json:"StartTime,omitempty" gorm:"column:start_time"`             // 开始时间
	EndTime     time.Time `json:"EndTime,omitempty" gorm:"column:end_time"`                 // 结束时间
}

// Round 轮次及打乱
type Round struct {
	ID        uint      `json:"ID" gorm:"primaryKey;column:id"`
	CreatedAt time.Time `json:"-" gorm:"autoCreateTime;column:created_at"`
	Name      string    `json:"Name,omitempty" gorm:"column:name"`
	ContestID uint      `json:"ContestID,omitempty" gorm:"column:contest_id"` // 所属比赛
	Project   string    `json:"Project,omitempty" gorm:"column:project"`      // 项目
	Number    int       `json:"Number,omitempty" gorm:"column:number"`        // 项目轮次
	Part      int       `json:"Part,omitempty" gorm:"column:part"`            // 该轮次第几份打乱
	Final     bool      `json:"Final,omitempty" gorm:"column:final"`          // 是否是最后一轮
	IsStart   bool      `json:"IsStart,omitempty" gorm:"column:is_start"`     // 是否已开始
	Upsets    string    `json:"-" gorm:"column:upsets"`                       // 打乱 UpsetDetail
	UpsetsVal []string  `json:"UpsetsVal,omitempty" gorm:"-"`                 // 打乱 UpsetDetail 实际内容
}

// Player 选手表
type Player struct {
	ID        uint      `json:"ID" gorm:"primaryKey;column:id"`
	CreatedAt time.Time `json:"-" gorm:"autoCreateTime;column:created_at"`

	Name       string `json:"Name" gorm:"unique;not null;column:name"` // 选手名
	WcaID      string `json:"WcaID,omitempty" gorm:"column:wca_id"`    // 选手WcaID，用于查询选手WCA的成绩
	ActualName string `json:"ActualName,omitempty" gorm:"actual_name"` // 真实姓名

	TitlesVal []string `json:"TitlesVal,omitempty" gorm:"-"`
	//DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除
}

// PlayerUser 选手用户表
type PlayerUser struct {
	ID           uint      `json:"ID" gorm:"primaryKey;column:id"`
	CreatedAt    time.Time `json:"-" gorm:"autoCreateTime;column:created_at"`
	IsAdmin      bool      `json:"isAdmin"`      // 是否为普通管理员
	IsSuperAdmin bool      `json:"isSuperAdmin"` // 是否为超级管理员

	LoginID  string `gorm:"column:login_id"`                  // 登录自定义ID
	PlayerID uint   `gorm:"unique;not null;column:player_id"` // 选手ID
	Password string `json:"-" gorm:""`                        // 密码 md5加密校验

	QQ         string `json:"-" gorm:"column:qq"`            // qq号
	QQBotUniID string `json:"-" gorm:"column:qq_bot_uni_id"` // qq 机器人ID
	WeChat     string `json:"-" gorm:"column:wechat"`        // 微信号
	Phone      string `json:"-" gorm:"column:phone"`         // 手机号
}
