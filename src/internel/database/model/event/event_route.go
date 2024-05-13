package event

type RouteType int

const (
	RouteTypeNot             RouteType = iota // 非比赛项目
	RouteType1rounds                          // "1_r"      // 单轮项目
	RouteType3roundsBest                      // "3_r_b"    // 三轮取最佳
	RouteType3roundsAvg                       // "3_r_a"    // 三轮取平均
	RouteType5roundsBest                      // "5_r_b"    // 五轮取最佳
	RouteType5roundsAvg                       // "5_r_a"    // 五轮取平均
	RouteType5RoundsAvgHT                     // "5_r_a_ht" // 五轮去头尾取平均
	RouteTypeRepeatedly                       // "ry"       // 单轮多次还原项目, 成绩1:还原数; 成绩2:尝试数; 成绩3:时间;
	RouteType2RepeatedlyBest                  // “2ry” 两轮多次尝试取最佳
	RouteType3RepeatedlyBest                  // "3ry"      // 三轮尝试多次还原项目 成绩1:还原数; 成绩2:尝试数; 成绩3:时间; 循环3次
)

func (r RouteType) RouteMap() RouteMap {
	return routeMaps[r]
}

func (r RouteType) String() string {
	return r.RouteMap().Name
}

type RouteMap struct {
	Name          string `json:"name"`          // 名称
	Repeatedly    bool   `json:"repeatedly"`    // 是否多轮还原项目
	RepeatedlyNum int    `json:"repeatedlyNum"` // 多轮还原项目的轮次
	Rounds        int    `json:"rounds"`        // 成绩数
	WithBest      bool   `json:"withBest"`      // 取最佳
	HeadToTailNum int    `json:"headToTailNum"` // 去头尾的数量
	NotComp       bool   `json:"notComp"`       // 是否非比赛项目
}

var routeMaps = map[RouteType]RouteMap{
	RouteTypeNot: RouteMap{
		Name:    "非比赛项目",
		NotComp: true,
	},
	RouteType1rounds: RouteMap{
		Name:     "一把取最佳",
		Rounds:   1,
		WithBest: true,
	},
	RouteType3roundsBest: RouteMap{
		Name:     "三把取最佳",
		Rounds:   3,
		WithBest: true,
	},
	RouteType3roundsAvg: RouteMap{
		Name:   "三把取平均",
		Rounds: 3,
	},
	RouteType5roundsBest: RouteMap{
		Name:          "五把取最佳",
		Rounds:        5,
		WithBest:      true,
		HeadToTailNum: 1,
	},
	RouteType5roundsAvg: RouteMap{
		Name:          "五把取平均",
		Rounds:        5,
		HeadToTailNum: 0,
	},
	RouteType5RoundsAvgHT: RouteMap{
		Name:          "五把去头尾取平均",
		Rounds:        5,
		HeadToTailNum: 1,
	},
	RouteTypeRepeatedly: RouteMap{
		Name:          "一把多次尝试",
		Repeatedly:    true,
		RepeatedlyNum: 1,
		Rounds:        3,
		WithBest:      true,
	},
	RouteType2RepeatedlyBest: RouteMap{
		Name:          "两把多次尝试取最佳",
		Repeatedly:    true,
		RepeatedlyNum: 2,
		Rounds:        6,
		WithBest:      true,
	},
	RouteType3RepeatedlyBest: RouteMap{
		Name:          "三把多次尝试取最佳",
		Repeatedly:    true,
		RepeatedlyNum: 3,
		Rounds:        9,
		WithBest:      true,
	},
}
