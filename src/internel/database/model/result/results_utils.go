package result

import (
	"fmt"
	"math"
	"sort"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

func UpdateOrgResult(
	in []float64, eventRoute event.RouteType,
	cutoff float64, cutoffNumber int,
	timeLimit float64,
) []float64 {
	for n := eventRoute.RouteMap().Rounds; len(in) < n; {
		if eventRoute.RouteMap().Repeatedly {
			in = append(in, 0, 0, DNS)
		} else {
			in = append(in, DNS)
		}
	}
	var out = make([]float64, eventRoute.RouteMap().Rounds)
	copy(out, in)

	if eventRoute.RouteMap().Repeatedly {
		return out
	}

	if timeLimit != 0 {
		for n := range out {
			if out[n] <= DNF {
				continue
			}
			if out[n] > timeLimit {
				out[n] = DNT
			}
		}
	}

	if cutoff != 0 && cutoffNumber != 0 {
		if cutoffNumber > len(out) {
			cutoffNumber = len(out)
		}
		var can bool
		for n := range out[:cutoffNumber] {
			if out[n] <= DNF {
				continue
			}
			if out[n] < cutoff {
				can = true
				break
			}
		}

		if !can {
			for n := cutoffNumber; n < len(out); n++ {
				out[n] = DNP
			}
		}
	}

	return out
}

func (c *Results) updateBestAndAvg() error {
	// 1. 拷贝基础数据并做
	for n := c.EventRoute.RouteMap().Rounds; len(c.Result) < n; {
		c.Result = append(c.Result, DNS)
	}

	cache := make([]float64, c.EventRoute.RouteMap().Rounds)
	copy(cache, c.Result)

	c.Best, c.Average, c.BestRepeatedlyTime = DNF, DNF, DNF

	// 2. 计次项目
	if c.EventRoute.RouteMap().Repeatedly {
		list := sortRepeatedly(getRepeatedlyList(cache, c.EventRoute.RouteMap().Rounds))
		if !list[0].D() {
			c.Best = list[0].N()
			c.BestRepeatedlyTime = list[0].Time
			c.BestRepeatedlyTry = list[0].Try
			c.BestRepeatedlyReduction = list[0].Reduction
		}
		// todo 多次取平均
		c.Average = 0
		return nil
	}

	// 3. 计时项目
	best, avg := getBestAndAvg(cache, c.EventRoute.RouteMap())
	c.Best = best
	c.Average = avg
	return nil
}

type repeatedly struct {
	Reduction float64
	Try       float64
	Time      float64
}

func getRepeatedlyList(in []float64, roundNum int) []repeatedly {
	if len(in) == 0 {
		return nil
	}
	list := make([]repeatedly, 0)

	for i := 0; i < len(in) && roundNum > 0; i, roundNum = i+3, roundNum-3 {
		list = append(list, repeatedly{in[i], in[i+1], in[i+2]})
	}
	return list
}

func (r repeatedly) D() bool {
	if r.Reduction < 2 {
		return true
	}
	if r.Try-r.Reduction > r.Reduction {
		return true
	}
	return false
}

// N  分数
func (r repeatedly) N() float64 {
	return r.Reduction - (r.Try - r.Reduction)
}

func sortRepeatedly(in []repeatedly) []repeatedly {
	sort.Slice(
		in, func(i, j int) bool {
			ir := in[i]
			ij := in[j]
			// 还原需要大于尝试数 还原数必须多于两把
			if ir.D() || ij.D() {
				return ij.D()
			}

			if ir.N() == ij.N() {
				return ir.Time < ij.Time
			}

			return ir.N() > ij.N()
		},
	)
	return in
}

func getBestAndAvg(results []float64, routeMap event.RouteMap) (best, avg float64) {
	best, avg = DNF, DNF

	// DNF
	d := 0
	for i := 0; i < len(results); i++ {
		if results[i] <= DNF {
			d++
		}
	}
	if d == len(results) || len(results) == 0 {
		return
	}

	// 排序
	sort.Slice(
		results, func(i, j int) bool {
			if results[i] <= DNF || results[j] <= DNF {
				return results[j] <= DNF
			}
			return results[i] <= results[j]
		},
	)

	// 最佳成绩
	best = results[0]
	if len(results) == 1 {
		return
	}

	// 去头尾平均, 大于一半可以不要平均了
	if d >= len(results)/2 || (len(results)-routeMap.HeadToTailNum*2) < 1 {
		return
	}
	avg = 0
	n := 0.0
	for idx, val := range results {
		if idx < routeMap.HeadToTailNum || idx >= len(results)-routeMap.HeadToTailNum {
			continue
		}
		if val <= DNF {
			avg = DNF
			return
		}
		avg += val
		n += 1
	}
	avg = math.Round((avg/n)*100) / 100.0
	return
}

func (c *Results) isBest(other Results) bool {
	if c.DBest() || other.DBest() {
		return !c.DBest()
	}

	if c.Best == other.Best {
		if c.EventRoute.RouteMap().Repeatedly {
			return c.BestRepeatedlyTime <= other.BestRepeatedlyTime
		}
		return c.Average <= other.Average
	}
	if c.EventRoute.RouteMap().Repeatedly {
		return c.Best >= other.Best
	}
	return c.Best <= other.Best
}

func (c *Results) isBestAvg(other Results) bool {
	if c.EventRoute.RouteMap().Repeatedly {
		return true
	}
	if c.DAvg() || other.DAvg() {
		return !c.DAvg()
	}
	if c.DAvg() && other.DAvg() {
		return c.IsBest(other)
	}
	if c.Average == other.Average {
		return c.isBest(other)
	}
	return c.Average <= other.Average
}

func SortResultWithBest(in []Results) {
	if len(in) == 0 {
		return
	}

	sort.Slice(in, func(i, j int) bool {
		return in[i].IsBest(in[j])
	})

	rom := in[0].EventRoute.RouteMap()

	in[0].Rank = 1
	prev := in[0]
	for i := 1; i < len(in); i++ {

		if (rom.Repeatedly && in[i].EqualRepeatedly(prev)) || (!rom.Repeatedly && in[i].Best == prev.Best) {
			in[i].Rank = prev.Rank
		} else {
			in[i].Rank = prev.Rank + 1
		}
		prev = in[i]
	}
	return
}

func SortResultWithAvg(in []Results) {
	if len(in) == 0 {
		return
	}

	sort.Slice(in, func(i, j int) bool {
		return in[i].IsBestAvg(in[j])
	})

	in[0].Rank = 1
	prev := in[0]
	for i := 1; i < len(in); i++ {
		if in[i].Average == prev.Average {
			in[i].Rank = prev.Rank
		} else {
			in[i].Rank = prev.Rank + 1
		}
		prev = in[i]
	}
	return
}

func SortResult(in []Results) {
	if len(in) <= 1 {
		return
	}

	rom := in[0].EventRoute.RouteMap()
	sort.Slice(
		in, func(i, j int) bool {
			if rom.WithBest {
				return in[i].isBest(in[j])
			}
			return in[i].isBestAvg(in[j])
		},
	)

	in[0].Rank = 1
	prev := in[0]
	for i := 1; i < len(in); i++ {
		if rom.WithBest {
			if in[i].Best == prev.Best {
				in[i].Rank = prev.Rank
				continue
			}
			in[i].Rank = i
			prev = in[i]
			continue
		}

		// 前面已经排序好了，只需要给同分即可
		if in[i].Average == prev.Average && in[i].Best == prev.Best {
			in[i].Rank = prev.Rank
			continue
		}
		in[i].Rank = i
		prev = in[i]
	}
}

func TimeParser(in float64) string {

	switch in {
	case DNF:
		return "DNF"
	case DNS:
		return "DNS"
	case DNP:
		return "DNP"
	case DNT:
		return "DNT"
	}

	// 判断是否超过2小时
	if in >= 2*60*60 {
		h := int(math.Floor(in) / 3600)  // 小时
		m := int(math.Floor(in)/60) % 60 // 分钟
		s := int(in) % 60                // 秒
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}

	if in < 60 {
		return fmt.Sprintf("%0.2f", in)
	}
	m := int(math.Floor(in) / 60)
	s := in - float64(m*60)

	ss := fmt.Sprintf("%0.2f", s)
	if s < 10 {
		ss = fmt.Sprintf("0%0.2f", s)
	}

	return fmt.Sprintf("%d:%s", m, ss)
}

func (c *Results) bestString() string {
	if c.DBest() {
		return ""
	}
	if c.EventRoute.RouteMap().Repeatedly {
		return fmt.Sprintf("%d/%d %s", int(c.BestRepeatedlyReduction), int(c.BestRepeatedlyTry), TimeParser(c.BestRepeatedlyTime))
	}
	if c.EventRoute.RouteMap().Integer {
		return fmt.Sprintf("%d", int(c.Best))
	}
	return TimeParser(c.Best)
}

func (c *Results) bestAvgString() string {
	if c.DAvg() {
		return ""
	}
	return TimeParser(c.Average)
}
