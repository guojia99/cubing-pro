package result

type PreResults struct {
	Results

	// 录入需要用到的字段:
	// CompetitionID, CompetitionName, Round, RoundNumber
	// PersonName, UserID, CubeID, Result
	// EventID, EventName, EventRoute
	// 成绩需要通过  Update() 进行处理

	ResultID     *uint  `gorm:"column:result_id"`     // 对应的ID
	CompsName    string `gorm:"column:comps_name"`    // 比赛名
	RoundName    string `gorm:"column:round_name"`    // 轮次名
	Recorder     string `gorm:"column:recorder"`      // 记录人
	Processor    string `gorm:"column:processor"`     // 处理人
	ProcessorID  uint   `gorm:"column:processor_id"`  // 处理人ID
	Finish       bool   `gorm:"column:finish"`        // 是否处理
	Detail       string `gorm:"column:detail"`        // 处理结果
	FinishDetail string `gorm:"column:finish_detail"` // 最终处理结果
	Source       string `gorm:"column:source"`        // 来源
}

const (
	DetailOk       = "ok"
	DetailNot      = "not"
	DetailWait     = "wait"
	DetailDoubtful = "doubtful"
	DetailTimeout  = "timeout" // 过期
)
