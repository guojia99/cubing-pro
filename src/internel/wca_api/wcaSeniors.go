package wca_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
	jsoniter "github.com/json-iterator/go"
)

const wcaSeniorsUrl = "https://wca-seniors.org/data/Senior_Rankings.js"
const extendKey = "rankings ="
const resetTime = time.Hour * 6

func getWcaSeniors() (*SeniorsData, error) {
	// 1. 创建 HTTP GET 请求
	resp, err := http.Get(wcaSeniorsUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // 确保关闭响应体

	// 2. 读取响应体内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 3. 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get code: %d", resp.StatusCode)
	}

	// 4. 处理响应数据：移除 extendKey 前缀（如果存在）
	responseStr := string(body)
	if strings.Contains(responseStr, extendKey) {
		responseStr = strings.Replace(responseStr, extendKey, "", 1)
	}

	// 5. 解析 JSON 到结构体
	var out SeniorsData
	err = jsoniter.Unmarshal([]byte(responseStr), &out)
	if err != nil {
		return nil, err
	}

	// 6. 后处理数据（如填充额外信息）
	return fillSeniorPersonData(&out), nil
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
var cacheFilePath = "wca_seniors_cache.json"

func updateCacheData() {
	log.Infof("start update cache data")
	d, err := getWcaSeniors()
	if err != nil {
		log.Errorf("update wca seniors CacheData error: %v", err)
		return
	}
	log.Infof("cache data update ok %s", d.Refreshed)
	cacheData = d

	// 保存到本地文件
	if err = saveCacheToFile(d); err != nil {
		log.Errorf("failed to save cache to file: %v", err)
	}
}

func loadCacheFromLocal() error {
	file, err := ioutil.ReadFile(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("cache file does not exist, will fetch from remote")
			return err
		}
		log.Errorf("failed to read cache file: %v", err)
		return err
	}

	var data SeniorsData
	if err = json.Unmarshal(file, &data); err != nil {
		log.Errorf("failed to unmarshal cache data: %v", err)
		return err
	}

	// 检查时间是否超过1天
	if isCacheExpired(data.Refreshed) {
		log.Info("cache is expired, will fetch fresh data")
		return err
	}

	cacheData = &data
	log.Info("loaded cache from local file successfully")
	return nil
}

func saveCacheToFile(data *SeniorsData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(cacheFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(cacheFilePath, jsonData, 0644)
}

func isCacheExpired(refreshedTime string) bool {
	if refreshedTime == "" {
		return true
	}

	parsedTime, err := time.Parse("2006-01-02 15:04:05", refreshedTime)
	if err != nil {
		log.Errorf("failed to parse refreshed time: %v", err)
		return true
	}

	// 检查是否超过1天
	return time.Since(parsedTime) > 24*time.Hour
}

func init() {
	// 启动定时更新
	go func() {
		// 先尝试从本地加载
		loadCacheFromLocal()

		// 如果本地加载失败或过期，则获取远程数据
		if cacheData == nil || isCacheExpired(cacheData.Refreshed) {
			updateCacheData()
		}

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

func GetSeniorsWithEventsAndGroup(country []string, age int, events []string) (BestSeniorValue, map[string][]SeniorPersonValue, error) {
	if cacheData == nil {
		return BestSeniorValue{}, nil, errors.New("seniors cache data is empty")
	}

	var ps = make(map[string][]SeniorPersonValue)

	var bestSingleCache = make(map[string][]SeniorRank)
	var bestAvgCache = make(map[string][]SeniorRank)
	var bv = BestSeniorValue{
		Single:  make(map[string]SeniorRank),
		Average: make(map[string]SeniorRank),
	}

	var checkIsInCountry = func(iso2 string) bool {
		if len(country) == 0 {
			return true
		}
		return slices.Contains(country, iso2)
	}

	for _, event := range events {
		ps[event] = make([]SeniorPersonValue, 0)
		bestSingleCache[event] = make([]SeniorRank, 0)
		bestAvgCache[event] = make([]SeniorRank, 0)

		for _, p := range cacheData.PersonMap {
			if !checkIsInCountry(p.Country) {
				continue
			}
			if _, ok := p.Single[age]; !ok {
				continue
			}
			if _, ok := p.Single[age][event]; !ok {
				continue
			}
			ps[event] = append(ps[event], p)
			bestSingleCache[event] = append(bestSingleCache[event], p.Single[age][event])
			if _, ok := p.Average[age]; !ok {
				continue
			}
			if _, ok := p.Average[age][event]; !ok {
				continue
			}
			bestAvgCache[event] = append(bestAvgCache[event], p.Average[age][event])
		}

		sort.Slice(bestSingleCache[event], func(i, j int) bool {
			return bestSingleCache[event][i].Rank < bestSingleCache[event][j].Rank
		})
		sort.Slice(bestAvgCache[event], func(i, j int) bool {
			return bestAvgCache[event][i].Rank < bestAvgCache[event][j].Rank
		})
		if len(bestSingleCache[event]) > 0 {
			bv.Single[event] = bestSingleCache[event][0]
		}
		if len(bestAvgCache[event]) > 0 {
			bv.Average[event] = bestAvgCache[event][0]
		}
	}

	return bv, ps, nil
}
