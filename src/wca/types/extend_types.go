package types

type ResultProportionEstimationType string

const (
	ResultProportionEstimationTypeCube333AndOh = ResultProportionEstimationType("333-333oh")
	ResultProportionEstimationTypeBigCube      = ResultProportionEstimationType("bigcube")
	ResultProportionEstimationTypeBLD          = ResultProportionEstimationType("bld")
)

var ResultProportionEstimationMap = map[ResultProportionEstimationType][]string{
	ResultProportionEstimationTypeCube333AndOh: {"333", "333oh"},
	ResultProportionEstimationTypeBigCube:      {"444", "555", "666", "777"},
	ResultProportionEstimationTypeBLD:          {"333bf", "444bf", "555bf"},
}

// ProportionEstimationSegment 表示按「锚点项目成绩」分段后，该段内顶尖选手群体在各项目上的相对比例。
// 比例定义为：t_event / t_anchor（均为同一选手、近期尝试的中位数），段内再对各选手比例取中位数以抗异常值。
type ProportionEstimationSegment struct {
	AnchorMin int                `json:"anchor_min"` // 锚点成绩下界（百分之一秒，含）
	AnchorMax int                `json:"anchor_max"` // 锚点成绩上界（百分之一秒，含）
	NPersons  int                `json:"n_persons"`  // 落入该段的选手数
	Ratio     map[string]float64 `json:"ratio"`      // event_id -> t_event/t_anchor；锚点项目自身比率为 1（可不出现）
}

// ProportionCurveSample 曲线采样点（便于前端绘图）；时间为秒（float），与注释示例一致。
type ProportionCurveSample struct {
	AnchorSec float64            `json:"anchor_sec"` // 锚点项目成绩（秒）
	Estimates map[string]float64 `json:"estimates"`  // 各项目估计成绩（秒）
}

// ResultProportionEstimationResult 多项目成绩比例静态估计的输出。
type ResultProportionEstimationResult struct {
	Persons []string `json:"persons"`

	// Events 与 ResultProportionEstimationMap 中顺序一致；Events[0] 为锚点项目（推断时的已知项）。
	Events []string `json:"events"`

	// Segments 分段比例表：相近锚点水平的选手归为一组，组内统计各项目相对锚点的中位比例。
	Segments []ProportionEstimationSegment `json:"segments"`

	// GlobalRatio 全体样本上的全局中位比例（锚点过极端或分段外推时使用）。
	GlobalRatio map[string]float64 `json:"global_ratio"`

	// CurveSamples 在观测锚点范围内的采样曲线（分段比例线性插值后换算为秒）。
	CurveSamples []ProportionCurveSample `json:"curve_samples"`

	// SampleCount 参与拟合的有效选手数（各项目均有至少一次有效尝试）。
	SampleCount int `json:"sample_count"`
}
