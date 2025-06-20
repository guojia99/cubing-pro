package job

import (
	"fmt"
	"path"
	"sync"

	"github.com/guojia99/cubing-pro/src/internel/database/model/system"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"gorm.io/gorm"
)

const (
	DiyRankingsKey  = "diy_rankings"
	DiyCubingProKey = "diy_rankings_cubing_pro"
	urlFormat       = "https://www.worldcubeassociation.org/persons/%s" // 2017XUYO01

	wcaUrlFormat = "https://www.worldcubeassociation.org/api/v0/persons/%s/results" // 2017XUYO01
)

type (
	EventData struct {
		Event          string `json:"event"`
		CountryRank    string `json:"country_rank"`
		ContinentRank  string `json:"continent_rank"`
		WorldRank      string `json:"world_rank"`
		Single         string `json:"single"`
		Average        string `json:"average"`
		WorldRank2     string `json:"world_rank_2"`
		ContinentRank2 string `json:"continent_rank_2"`
		CountryRank2   string `json:"country_rank_2"`
	}

	PersonResult struct {
		WCA    string
		Name   string
		Events []EventData
	}

	Persons struct {
		WcaId string `gorm:"column:wca_id"`
		Name  string `gorm:"column:name"`
	}

	Results struct {
		EventId    string `json:"eventId"`
		Best       int    `json:"best"`
		BestStr    string `json:"bestStr"`
		Average    int    `json:"average"`
		AverageStr string `json:"averageStr"`
		PersonName string `json:"personName"`
		PersonId   string `json:"personId"`
	}

	PersonBestResults struct {
		PersonName string             `json:"PersonName"`
		Best       map[string]Results `json:"Best"`
		Avg        map[string]Results `json:"Avg"`
	}
	WcaResult struct {
		BestRank        int    `json:"BestRank"`
		BestStr         string `json:"BestStr"`
		BestPersonName  string `json:"BestPersonName"`
		BestPersonWCAID string `json:"BestPersonWCAID"`
		AvgRank         int    `json:"AvgRank"`
		AvgStr          string `json:"AvgStr"`
		AvgPersonName   string `json:"AvgPersonName"`
		AvgPersonWCAID  string `json:"AvgPersonWCAID"`
	}
)
type UpdateDiyRankings struct {
	DB *gorm.DB

	one sync.Once
}

func (u *UpdateDiyRankings) Name() string { return "UpdateDiyRankings" }

func (u *UpdateDiyRankings) updateWCAResult() error {
	var keys []string
	if err := system.GetKeyJSONValue(u.DB, DiyRankingsKey, &keys); err != nil {
		return err
	}
	for _, key := range keys {
		var wcaKeys []string
		if err := system.GetKeyJSONValue(u.DB, key, &wcaKeys); err != nil {
			continue
		}

		dataKey := path.Join(DiyRankingsKey, key, "data")
		// 更换为WCA的逻辑代码
		data := u.apiGetSortResult(wcaKeys)
		_ = system.SetKeyJSONValue(u.DB, dataKey, data, "")
		fmt.Printf("[UpdateDiyRankings] 更新数据 %s\n", key)
	}
	return nil
}

func (u *UpdateDiyRankings) updateCubingPro() error {
	var users []user2.User
	u.DB.Find(&users)

	var wcaId []string
	for _, uu := range users {
		if uu.WcaID != "" {
			wcaId = append(wcaId, uu.WcaID)
		}
	}

	return system.SetKeyJSONValue(u.DB, DiyCubingProKey, wcaId, "网站成员榜单")
}

func (u *UpdateDiyRankings) updateCubingProKey() {
	var keys []string
	_ = system.GetKeyJSONValue(u.DB, DiyRankingsKey, &keys)
	keys = append(keys, DiyCubingProKey)
	keys = utils.RemoveDuplicates(keys)

	_ = system.SetKeyJSONValue(u.DB, DiyRankingsKey, keys, "自定义WCAID榜单列表")
}

func (u *UpdateDiyRankings) Run() error {
	u.one.Do(u.updateCubingProKey)
	_ = u.updateCubingPro()
	return u.updateWCAResult()
}
