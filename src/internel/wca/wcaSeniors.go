package wca

import (
	"errors"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
	jsoniter "github.com/json-iterator/go"
)

const wcaSeniorsUrl = "https://wca-seniors.org/data/Senior_Rankings.js"
const extendKey = "rankings ="
const resetTime = time.Hour * 6

func getWcaSeniors() (*SeniorsData, error) {
	var out *SeniorsData

	resp, err := utils.HTTPRequestFull("GET", wcaSeniorsUrl, nil, nil, nil)
	if err != nil {
		return out, err
	}

	d := string(resp.Body)

	err = jsoniter.UnmarshalFromString(strings.Replace(d, extendKey, "", 1), &out)
	if err != nil {
		return out, err
	}
	return fillSeniorPersonData(out), nil
}
func fillSeniorPersonData(data *SeniorsData) *SeniorsData {
	// 构建国家、洲映射
	countryMap := make(map[string]*SeniorCountry)
	for i := range data.Countries {
		c := &data.Countries[i]
		countryMap[c.ID] = c
	}

	continentMap := make(map[string]*SeniorContinent)
	for i := range data.Continents {
		ct := &data.Continents[i]
		continentMap[ct.ID] = ct
	}

	// 初始化 PersonMap
	data.PersonMap = make(map[string]SeniorPersonValue)
	for _, p := range data.Persons {
		countryName := ""
		continent := ""
		if c, ok := countryMap[p.Country]; ok {
			countryName = c.Name
			continent = c.Continent
		}
		data.PersonMap[p.Id] = SeniorPersonValue{
			SeniorPerson: p,
			CountryName:  countryName,
			Continent:    continent,
			Single:       make(map[int]map[string]SeniorRank),
			Average:      make(map[int]map[string]SeniorRank),
		}
	}

	// 遍历 event → ranking → ranks
	for _, ev := range data.Events {
		for _, ranking := range ev.Rankings {
			// 维护洲/国家计数器
			contCounters := make(map[string]int)
			countryCounters := make(map[string]int)

			for _, r := range ranking.Ranks {
				// 填充 type/age
				r.Type = ranking.Type
				r.Age = ranking.Age

				// WR = Rank
				r.WR = r.Rank

				// 找出选手所属 country / continent
				pv, ok := data.PersonMap[r.Id]
				if !ok {
					// 如果选手没出现在 persons 列表，创建占位
					pv = SeniorPersonValue{
						SeniorPerson: SeniorPerson{
							Id:      r.Id,
							Country: "",
						},
						Single:  make(map[int]map[string]SeniorRank),
						Average: make(map[int]map[string]SeniorRank),
					}
				}

				countryCode := pv.Country
				continentCode := pv.Continent
				if countryCode != "" {
					if c, ok := countryMap[countryCode]; ok {
						continentCode = c.Continent
					}
				}

				// 计算 CR/NR
				if continentCode == "" {
					continentCode = "__unknown__"
				}
				if countryCode == "" {
					countryCode = "__unknown__"
				}
				contCounters[continentCode]++
				countryCounters[countryCode]++
				r.CR = contCounters[continentCode]
				r.NR = countryCounters[countryCode]

				// 存入 PersonMap：先按 age，再按 eventId
				if r.Type == "single" {
					if pv.Single[r.Age] == nil {
						pv.Single[r.Age] = make(map[string]SeniorRank)
					}
					pv.Single[r.Age][ev.Id] = r
				} else {
					if pv.Average[r.Age] == nil {
						pv.Average[r.Age] = make(map[string]SeniorRank)
					}
					pv.Average[r.Age][ev.Id] = r
				}

				data.PersonMap[r.Id] = pv
			}
		}
	}

	return data
}

var cacheData *SeniorsData

func updateCacheData() {
	d, err := getWcaSeniors()
	if err != nil {
		log.Errorf("update wca seniors CacheData error: %v", err)
		return
	}
	cacheData = d
}

func init() {
	updateCacheData()
	go func() {
		ticker := time.NewTicker(resetTime)
		for {
			select {
			case <-ticker.C:
				updateCacheData()
			}
		}
	}()
}

func GetSeniorsPerson(wcaID string) (*SeniorPersonValue, error) {
	if cacheData == nil {
		return nil, errors.New("seniors cache data is empty")
	}

	p, ok := cacheData.PersonMap[wcaID]
	if !ok {
		return nil, errors.New("seniors person not found")
	}
	return &p, nil
}
