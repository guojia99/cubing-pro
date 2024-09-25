package job

import (
	"fmt"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"gorm.io/gorm"
)

type RecordUpdateJob struct {
	DB *gorm.DB
}

func (c *RecordUpdateJob) Name() string {
	return "RecordUpdateJob"
}

func (c *RecordUpdateJob) getRecords(where string, typ string) []result.Record {
	// 1. 分段获取所有比赛的成绩
	//      - 一次获取20个比赛的id
	//      - 通过这20个比赛id，拉取比赛成绩数据
	// 2. 每个比赛只保留一份（可并列）最佳成绩
	// 3. 每次循环都检查当前成绩是否比上次的好。
	// 4. 结束时，删除所有记录，然后重新写入。
	var records []result.Record
	var nowBest = make(map[string]result.Results) // key is eventName
	var nowAvg = make(map[string]result.Results)  // key is eventName

	addRecord := func(best bool, r result.Results, comp competition.Competition) {
		record := result.Record{
			Type:        typ,
			EventId:     r.EventID,
			EventRoute:  r.EventRoute,
			ResultId:    r.ID,
			UserId:      r.UserID,
			UserName:    r.PersonName,
			CompsId:     r.CompetitionID,
			CompsName:   comp.Name,
			CompsGenre:  comp.Genre,
			ThisResults: r.ResultJSON,
		}
		if best {
			if r.EventRoute.RouteMap().Repeatedly {
				s := r.BestString()
				record.Repeatedly = &s
			} else {
				record.Best = &r.Best
			}
		} else {
			record.Average = &r.Average
		}
		records = append(records, record)
	}

	updateCompWithPage := func(page int) error {
		// 1. 查询比赛
		var comps []competition.Competition
		if err := c.DB.Model(&competition.Competition{}).Where(where).Limit(20).Offset(page * 20).Find(&comps).Error; err != nil {
			return err
		}
		if len(comps) == 0 {
			return fmt.Errorf("end comp")
		}
		var compsIds []uint
		for _, comp := range comps {
			compsIds = append(compsIds, comp.ID)
		}

		// 2. 查询所有比赛的成绩
		var results []result.Results
		if err := c.DB.Where("comp_id in ?", compsIds).Find(&results).Error; err != nil {
			return err
		}

		// 3. 给成绩按照比赛做分类
		var resultWithComps = make(map[uint][]result.Results)
		for _, compId := range compsIds {
			resultWithComps[compId] = make([]result.Results, 0)
		}
		for _, r := range results {
			resultWithComps[r.CompetitionID] = append(resultWithComps[r.CompetitionID], r)
		}

		// 4. 获取每场比赛最佳成绩
		for _, comp := range comps {
			compResults, ok := resultWithComps[comp.ID]
			if !ok || len(compResults) == 0 {
				continue
			}
			var withEventBest = make(map[string][]result.Results) // 一场比赛多个记录
			var withEventAvg = make(map[string][]result.Results)  // 一场比赛多个记录

			for _, r := range compResults {
				// 单次
				func() {
					if r.DBest() {
						return
					}
					if _, ok2 := withEventBest[r.EventID]; !ok2 {
						withEventBest[r.EventID] = []result.Results{r}
						return
					}
					best := withEventBest[r.EventID][0]

					if r.EventRoute.RouteMap().Repeatedly {
						if r.IsBest(best) {
							withEventBest[r.EventID] = []result.Results{r}
						}
						return
					}

					if best.Best == r.Best {
						withEventBest[r.EventID] = append(withEventBest[r.EventID], r)
						return
					}
					if r.Best < best.Best {
						withEventBest[r.EventID] = []result.Results{r}
					}
				}()

				// 平均
				if r.EventRoute.RouteMap().Repeatedly {
					continue
				}
				if r.DAvg() {
					continue
				}

				if _, ok2 := withEventAvg[r.EventID]; !ok2 {
					withEventAvg[r.EventID] = []result.Results{r}
					continue
				}
				avg := withEventAvg[r.EventID][0]
				if r.Average == avg.Average {
					withEventAvg[r.EventID] = append(withEventAvg[r.EventID], r)
					continue
				}
				if r.Average < avg.Average {
					withEventAvg[r.EventID] = []result.Results{r}
				}
			}

			// 最佳成绩
			for key, val := range withEventBest {
				// 不存在时
				// 成绩比以前的好时
				if b, ok3 := nowBest[key]; !ok3 || (b.EventRoute.RouteMap().Repeatedly && val[0].IsBest(b)) || val[0].Best < b.Best {
					for _, v := range val {
						addRecord(true, v, comp)
					}
					nowBest[val[0].EventID] = val[0]
					continue
				}
			}

			// 平均成绩
			for key, val := range withEventAvg {
				if _, ok3 := nowAvg[key]; !ok3 || val[0].Average < nowAvg[key].Average {
					for _, v := range val {
						addRecord(false, v, comp)
					}
					nowAvg[key] = val[0]
				}
			}
		}
		return nil
	}

	for page := 0; ; page++ {
		if err := updateCompWithPage(page); err != nil {
			break
		}
	}
	return records
}

func (c *RecordUpdateJob) Run() error {
	// base records
	records := c.getRecords("", result.RecordTypeWithCubingPro)
	var baseRecordsMap = make(map[string]result.Record)
	for _, record := range records {
		baseRecordsMap[record.Key()] = record
	}

	// groups
	var groups []competition.CompertionGroup
	c.DB.Find(&groups)
	for _, group := range groups {
		rs := c.getRecords(fmt.Sprintf("group_id = %d", group.ID), result.RecordTypeWithGroup)
		for _, record := range rs {
			if _, ok := baseRecordsMap[record.Key()]; ok {
				continue
			}
			records = append(records, record)
		}
	}

	//// base
	//records := c.getRecords("", result.RecordTypeWithCubingPro)
	//var groups []competition.CompertionGroup
	//
	//// group
	//c.DB.Find(&groups)
	//for _, group := range groups {
	//	rs := c.getRecords(fmt.Sprintf("group_id = %d", group.ID), result.RecordTypeWithGroup)
	//	records = append(records, rs...)
	//}

	// todo 如果GR破了CR，则这个CR也要删除

	fmt.Printf("[Record] update record = %d", len(records))
	if err := c.DB.Where("1 = 1").Delete(&result.Record{}).Error; err != nil {
		return err
	}

	if err := c.DB.Save(&records).Error; err != nil {
		return err
	}
	return nil
}
