package result

import (
	"sort"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

func (c *Results) updateBestAndAvg() error {
	// 1. 拷贝基础数据并做
	for n := c.EventRoute.RouteMap().Rounds; len(c.Result) < n; {
		c.Result = append(c.Result, DNS)
	}
	cache := make([]float64, len(c.Result))
	copy(cache, c.Result)

	c.Best, c.Average, c.BestRepeatedly = DNF, DNF, DNF

	// 2. 计次项目
	if c.EventRoute.RouteMap().Repeatedly {
		list := sortRepeatedly(getRepeatedlyList(cache))
		if !list[0].D() {
			c.Best = list[0].N()
			c.BestRepeatedly = list[0].Time
		}
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

func getRepeatedlyList(in []float64) []repeatedly {
	list := make([]repeatedly, 0)
	for i := 0; i < len(in); i += 3 {
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

	// 去头尾平均
	if d == len(results) || (len(results)-routeMap.HeadToTailNum*2) < 1 {
		return
	}
	for idx, val := range results {
		if idx < routeMap.HeadToTailNum || idx > len(results)-routeMap.HeadToTailNum {
			continue
		}
		avg += val
	}
	return
}

func (c *Results) isBest(other Results) bool {
	if c.EventRoute.RouteMap().Repeatedly {
		// blind cube special rules:
		// - the result1 is number of successful recovery.
		// - the result2 is number of attempts to recover.
		// - the result3 is use times, (max back row).
		// - sort priority： r1 > r2 > r3
		// - like: if r1 and r2 equal, the best r3 is rank the top.

		if c.DBest() || other.DBest() {
			return !c.DBest()
		}
		if c.Best == other.Best {
			// 成绩3
			return c.BestRepeatedly < other.BestRepeatedly
		}
		return c.Best > other.Best
	}

	if c.DBest() || other.DBest() {
		return !c.DBest()
	}
	if c.Best == other.Best {
		return c.Average < other.Average
	}
	return c.Best < other.Best
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
	return c.Average < other.Average
}

// SortResultsByBest 先看best 再看平均
func SortResultsByBest(in []Results) {
	if len(in) <= 1 {
		return
	}
	sort.Slice(
		in, func(i, j int) bool {
			return in[i].isBest(in[j])
		},
	)
}

func SortResultsByAvg(in []Results) {
	if len(in) <= 1 {
		return
	}
	sort.Slice(
		in, func(i, j int) bool {
			return in[i].isBestAvg(in[j])
		},
	)
}

func SortResultsAndUpdateRank(rt event.RouteType, in []Results) {
	if len(in) <= 1 {
		return
	}

	//if rt.WithBest() {
	//	SortResultsByBest(in)
	//}
	//if !rt.WithBest() {
	//
	//}

	// 1. 进行排序
	// 2. 给rank值
}
