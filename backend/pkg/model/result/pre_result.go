package result

type PreResults struct {
	Results

	CompsName    string `gorm:"column:comps_name"`    // 比赛名
	RoundName    string `gorm:"column:round_name"`    // 轮次名
	Recorder     string `gorm:"column:recorder"`      // 记录人
	Processor    string `gorm:"column:processor"`     //  处理人ID
	Finish       bool   `gorm:"column:finish"`        // 是否处理
	FinishDetail string `gorm:"column:finish_detail"` // 处理结果
	Source       string `gorm:"column:source"`        // 来源
}
