package algs

const (
	algsInfoFile         = "algs_info.json"
	algsetsInfoFile      = "algsets_info.json"
	algImagesInfoFile    = "combined.json"
	algGroupFile         = "groups_info.json"
	algScramblesInfoFile = "scrambles.json"
)

type (
	TrainerAlgorithm struct {
		Algs     []string    `json:"a"`
		Name     interface{} `json:"name"`
		Group    string      `json:"group"`
		Algset   string      `json:"algset"`
		Scramble string      `json:"scramble"`
	}
)

type AlgorithmConfigWithTrainer struct {
	Name string `json:"name"`

	AlgsInfo    map[string]TrainerAlgorithm `json:"algs_info"`    // key is 1 algs_info.json
	AlgsetsInfo map[string][]string         `json:"algsets_info"` // 大组 algsets_info.json
	SetKeys     []string                    `json:"set_keys"`     // 按顺序的json出key
	GroupsInfo  map[string][]int            `json:"groups_info"`  // 小组 groups_info.json

	Images    map[string]string   `json:"images"`    // combined.json
	Scrambles map[string][]string `json:"scrambles"` //  scrambles.json
}
