package algs

// Algorithm 单个公式
type Algorithm struct {
	Name      string   `json:"name"`
	Algs      []string `json:"algs"`
	Image     string   `json:"image"` // svg
	Scrambles []string `json:"scrambles"`
}

// AlgorithmGroup 一个大类里面的分组, 如EG1-H
type AlgorithmGroup struct {
	Name       string      `json:"name"`
	Algorithms []Algorithm `json:"algs"`
}

// AlgorithmSet 一个汇总的大类 如EG1, LEG
type AlgorithmSet struct {
	Name            string           `json:"name"`
	AlgorithmGroups []AlgorithmGroup `json:"groups"`
	GroupsKeys      []string         `json:"groups_keys"`
}

// AlgorithmClass 一个汇总的 公式集合, 如EG, FH, TEG
type AlgorithmClass struct {
	Name    string         `json:"name"`
	Sets    []AlgorithmSet `json:"sets"`
	SetKeys []string       `json:"setKeys"`
}

// CubeAlgorithms 一种魔方 如222
type CubeAlgorithms struct {
	Cube      string           `json:"cube"`
	ClassList []AlgorithmClass `json:"class"`
	ClassKeys []string         `json:"class_keys"`
}
