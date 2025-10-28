package wca_api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/models"
	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/utils"
	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
)

const wcaPersonUrlFormat = "https://www.worldcubeassociation.org/api/v0/persons/%s"

func ApiGetWCAPerson(wcaID string) (*PersonProfile, error) {
	wcaID = strings.ToUpper(wcaID)
	var resp = &PersonProfile{}
	if err := utils2.HTTPRequestWithJSON(http.MethodGet, fmt.Sprintf(wcaPersonUrlFormat, wcaID), nil, nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func GetWCAPersonResult(wcaID string) (*models.PersonBestResults, error) {
	pf, err := ApiGetWCAPerson(wcaID)
	if err != nil {
		return nil, err
	}

	var out = &models.PersonBestResults{
		PersonName: pf.Person.Name,
		WCAID:      pf.Person.WcaId,
		Best:       make(map[string]models.Results),
		Avg:        make(map[string]models.Results),
		MedalCount: models.MedalCount{
			Gold:   pf.Medals.Gold,
			Silver: pf.Medals.Silver,
			Bronze: pf.Medals.Bronze,
			Total:  pf.Medals.Total,
		},
		RecordCount: models.RecordCount{
			National:    pf.Records.National,
			Continental: pf.Records.Continental,
			World:       pf.Records.World,
			Total:       pf.Records.Total,
		},
		CompetitionCount: pf.CompetitionCount,
	}

	for key, val := range pf.PersonalRecords {
		if val.Single.Id == 0 {
			continue
		}
		out.Best[key] = models.Results{
			EventId:       val.Single.EventId,
			Best:          val.Single.Best,
			BestStr:       utils.ResultsTimeFormat(val.Single.Best, val.Single.EventId),
			PersonName:    pf.Person.Name,
			PersonId:      pf.Person.WcaId,
			WorldRank:     val.Single.WorldRank,
			ContinentRank: val.Single.ContinentRank,
			CountryRank:   val.Single.CountryRank,
		}

		if val.Average.Id == 0 {
			continue
		}
		out.Avg[key] = models.Results{
			EventId:       val.Average.EventId,
			Average:       val.Average.Best,
			AverageStr:    utils.ResultsTimeFormat(val.Average.Best, val.Average.EventId),
			PersonName:    pf.Person.Name,
			PersonId:      pf.Person.WcaId,
			WorldRank:     val.Average.WorldRank,
			ContinentRank: val.Average.ContinentRank,
			CountryRank:   val.Average.CountryRank,
		}
	}

	return out, nil
}
