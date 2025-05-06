package cubing

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/patrickmn/go-cache"
)

type WcaCompetition struct {
	Id                string    `json:"id"`
	Name              string    `json:"name"`
	Venue             string    `json:"venue"`
	RegistrationOpen  time.Time `json:"registration_open"`
	RegistrationClose time.Time `json:"registration_close"`
	StartDate         string    `json:"start_date"`
	EndDate           string    `json:"end_date"`
	ShortDisplayName  string    `json:"short_display_name"`
	City              string    `json:"city"`
	CountryIso2       string    `json:"country_iso2"`
	EventIds          []string  `json:"event_ids"`
	LatitudeDegrees   float64   `json:"latitude_degrees"`
	LongitudeDegrees  float64   `json:"longitude_degrees"`
	AnnouncedAt       time.Time `json:"announced_at"`
	Class             string    `json:"class"`
}

const wcaCompUrl = "https://www.worldcubeassociation.org/api/v0/competition_index"

const wcaInfoUrlFormat = "https://www.worldcubeassociation.org/api/v0/competitions/%s/wcif/public"

var wcaCitys = map[string]string{
	"中国":   "CN", // 中国
	"中国香港": "HK", // 香港
	"韩国":   "KR", // 韩国
	"马来西亚": "MY", // 马来
	"新加坡":  "SG", // 新加坡
	"越南":   "VN", // 越南
	"泰国":   "TH", // 泰国
	"日本":   "JP", // 日本
	"印尼":   "ID", // 印度尼西亚
	"英国":   "GB", // 英国
	"菲律宾":  "PH", // 菲律宾
}

func GetWcaComps(city string) []WcaCompetition {
	query := url.Values{}
	query.Set("country_iso2", city)
	query.Set("include_cancelled", "false")
	query.Set("ongoing_and_future", time.Now().AddDate(0, 0, -1).Format("2006-01-02"))
	query.Set("page", "1")

	// 构造完整 URL
	fullUrl := fmt.Sprintf("%s?%s", wcaCompUrl, query.Encode())

	// 发送 GET 请求
	resp, err := http.Get(fullUrl)
	if err != nil {
		fmt.Println("请求失败:", err)
		return []WcaCompetition{}
	}
	defer resp.Body.Close()

	// 解析 JSON 响应
	var result []WcaCompetition
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return []WcaCompetition{}
	}
	// 打印结果
	fmt.Println("请求结果:", result)
	return result
}

func GetAllWcaComps() map[string][]WcaCompetition {
	var out = make(map[string][]WcaCompetition)
	for key, ct := range wcaCitys {
		cps := GetWcaComps(ct)
		if len(cps) == 0 {
			continue
		}
		out[key] = cps
		log.Printf("%s - %d", key, len(cps))
	}
	return out
}

var wcaInfoCache = cache.New(time.Second*30, time.Minute)

func GetWcaInfo(id string) WCAInfo {
	if resp, ok := wcaInfoCache.Get(id); ok {
		return resp.(WCAInfo)
	}

	var wcaInfo WCAInfo
	err := utils.HTTPRequestWithJSON("GET", fmt.Sprintf(wcaInfoUrlFormat, id), nil, nil, nil, &wcaInfo)
	if err != nil {
		return wcaInfo
	}
	log.Printf("[Debug] get WCA Info %s\n", id)
	wcaInfoCache.Set(id, wcaInfo, time.Minute)
	return wcaInfo
}
