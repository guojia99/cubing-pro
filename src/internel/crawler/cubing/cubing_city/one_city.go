package cubing_city

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/mozillazg/go-pinyin"
)

// "https://ss.sxmfxh.com/comps/getCompByStatus?query={%%22c_status%%22:%%225,6,7,8%%22,%%22page%%22:1,%%22size%%22:100}"
const oneUrl = "https://ss.sxmfxh.com/comps/getCompByStatus?query={%%22c_status%%22:%%22%s%%22,%%22page%%22:%d,%%22size%%22:%d}"

type OneComp struct {
	CId          int         `json:"c_id"`
	CName        string      `json:"c_name"`
	CStatus      int         `json:"c_status"`
	IsCCA        int         `json:"is_CCA"`
	CLayer       int         `json:"c_layer"`
	CType        int         `json:"c_type"`
	CLinkId      int         `json:"c_link_id"`
	HostUId      int         `json:"host_u_id"`
	HostUName    string      `json:"host_u_name"`
	InputUId     string      `json:"input_u_id"`
	InputUName   string      `json:"input_u_name"`
	Province     string      `json:"province"`
	City         string      `json:"city"`
	District     string      `json:"district"`
	ProvinceId   int         `json:"province_id"`
	CityId       int         `json:"city_id"`
	DistrictId   int         `json:"district_id"`
	CAddress     string      `json:"c_address"`
	CDate        string      `json:"c_date"`
	CDays        int         `json:"c_days"`
	CGroup       string      `json:"c_group"`
	CEvent       string      `json:"c_event"`
	CFee         int         `json:"c_fee"`
	EventFee     string      `json:"event_fee"`
	EventRound   string      `json:"event_round"`
	EventFormart string      `json:"event_formart"`
	Capacity     int         `json:"capacity"`
	TimeLimit    string      `json:"time_limit"`
	RegBeginTime string      `json:"reg_begin_time"`
	RegEndTime   string      `json:"reg_end_time"`
	RegULink     string      `json:"reg_u_link"`
	RegVLink     string      `json:"reg_v_link"`
	InfoText     string      `json:"info_text"`
	Pic          string      `json:"pic"`
	Views        int         `json:"views"`
	CreateTime   time.Time   `json:"create_time"`
	UpdateTime   time.Time   `json:"update_time"`
	Longitude    string      `json:"longitude"`
	Latitude     string      `json:"latitude"`
	IsPayOnline  int         `json:"is_pay_online"`
	GroupCode    interface{} `json:"group_code"`
	BId          int         `json:"b_id"`
	Remark       string      `json:"remark"`
	Total        int         `json:"total"`
}

type OneResponse struct {
	Code int       `json:"code" :"Code"`
	Data []OneComp `json:"data" :"Data"`
}

func getOneCompLists() []OneComp {

	status := "5,6,7,8"
	size := 100
	page := 0
	total := 1000000

	var out []OneComp

	for {
		if size*page >= total {
			break
		}
		page += 1
		var resp OneResponse
		url := fmt.Sprintf(oneUrl, status, page, size)
		if err := utils.HTTPRequestWithJSON(http.MethodPost, url, nil, nil, nil, &resp); err != nil {
			fmt.Println(err)
			break
		}
		if len(resp.Data) > 0 {
			total = resp.Data[0].Total
		}
		out = append(out, resp.Data...)
	}
	return out
}

const zhongXiongKey = "中匈青少年国际魔方公开赛"

func chineseToPinyin(text string) string {
	a := pinyin.NewArgs()
	result := pinyin.Pinyin(text, a)
	pinyinStr := ""
	for _, s := range result {
		if len(s) > 0 {
			pinyinStr += s[0]
		}
	}
	return strings.ToUpper(string(pinyinStr[0])) + pinyinStr[1:]
}

func GetOneCityList() []string {
	list := getOneCompLists()
	var output []string

	for _, v := range list {
		re := regexp.MustCompile(`([\p{Han}]+)分站赛`)
		matches := re.FindAllStringSubmatch(v.CName, -1)
		if len(matches) == 0 {
			continue
		}
		//fmt.Println(matches[len(matches)-1][1])
		city := strings.ReplaceAll(matches[len(matches)-1][1], zhongXiongKey, "")
		output = append(output, chineseToPinyin(city))
	}

	return output
}
