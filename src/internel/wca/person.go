package wca

import (
	"fmt"
	"net/http"

	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
)

const wcaPersonSearchUrlFormat = "https://www.worldcubeassociation.org/persons?search=%s&order=asc&offset=0&limit=10&region=all"

type Person struct {
	Name              string `json:"name"`
	WcaId             string `json:"wca_id"`
	Country           string `json:"country"`
	CompetitionsCount int    `json:"competitions_count"`
	PodiumsCount      int    `json:"podiums_count"`
}

type ApiSearchPersonsResp struct {
	Total int      `json:"total"`
	Rows  []Person `json:"rows"`
}

func ApiSearchPersons(name string) (ApiSearchPersonsResp, error) {
	url := fmt.Sprintf(wcaPersonSearchUrlFormat, name)
	//path := fmt.Sprintf("/persons?search=%s&order=asc&offset=0&limit=10&region=all", name)
	var resp ApiSearchPersonsResp
	if err := utils2.HTTPRequestWithJSON(http.MethodGet, url, nil, map[string]interface{}{
		"Cache-Control":   "max-age=0, private, must-revalidate",
		"Content-Type":    "application/json; charset=utf-8",
		"Pragma":          "no-cache",
		"Accept":          "application/json",
		"Accept-Language": "zh-CN,zh-HK;q=0.9,zh;q=0.8,zh-TW;q=0.7,en;q=0.6",
		//":authority":      "www.worldcubeassociation.org",
		//":method":         "GET",
		//":path":           path,
		//":scheme":         "https",
		"Priority": "u=1, i",
		"Referer":  url,
	}, nil, &resp); err != nil {
		return ApiSearchPersonsResp{}, err
	}
	return resp, nil
}
