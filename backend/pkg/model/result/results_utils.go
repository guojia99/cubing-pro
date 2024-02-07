package result

import (
	"sort"
)

func (c *Results) updateBestAndAvg() error {
	// 1. 拷贝基础数据并做
	for n := c.EventRoute.N(); len(c.Result) < n; {
		c.Result = append(c.Result, DNS)
	}
	cache := make([]float64, len(c.Result))
	copy(cache, c.Result)

	c.Best, c.Average, c.BestRepeatedlyTime = DNF, DNF, DNF

	// 2. 计次项目
	switch c.EventRoute {
	case RouteTypeRepeatedly, RouteType3RepeatedlyBest:
		list := sortRepeatedly(getRepeatedlyList(cache))
		if !list[0].D() {
			c.Best = list[0].N()
			c.BestRepeatedlyTime = list[0].Time
		}
		c.Average = 0
		for _, val := range list {
			if val.D() {
				c.Average = DNF
				return nil
			}
			c.Average += val.N()
		}
		c.Average /= 3
		return nil
	}

	// 3. 计时项目
	best, avg, htAvg := getBestAndAvg(cache)
	c.Best = best
	c.Average = avg
	if c.EventRoute == RouteType5RoundsAvgHT {
		c.Average = htAvg
	}
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

func getBestAndAvg(results []float64) (best, avg float64, htAvg float64) {
	best, avg, htAvg = DNF, DNF, DNF

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

	if d >= 2 || d == len(results) {
		return
	}

	// 去头尾平均
	if len(results) >= 3 {
		for idx, val := range results {
			if idx == 0 || idx == len(results)-1 {
				continue
			}
			htAvg += val
		}
		htAvg /= float64(len(results) - 2)
	}

	// 平均
	if d == 0 {
		for _, val := range results {
			avg += val
		}
		avg /= float64(len(results))
	}
	return
}

func (c *Results) isBest(other Results) bool {
	switch c.EventRoute {
	case RouteTypeRepeatedly, RouteType3RepeatedlyBest:
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
			return c.BestRepeatedlyTime < other.BestRepeatedlyTime
		}
		return c.Best > other.Best
	default:
		if c.DBest() || other.DBest() {
			return !c.DBest()
		}
		if c.Best == other.Best {
			return c.Average < other.Average
		}
		return c.Best < other.Best
	}

}
func (c *Results) isBestAvg(other Results) bool {
	switch c.EventRoute {
	case RouteTypeRepeatedly, RouteType3RepeatedlyBest:
		return true
	default:
		if c.DAvg() || other.DAvg() {
			return !c.DAvg()
		}
		if c.DAvg() && other.DAvg() {
			return c.IsBest(other)
		}
		return c.Average < other.Average
	}
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

func SortResultsAndUpdateRank(rt RouteType, in []Results) {
	if len(in) == 0 {
		return
	}

	if rt.WithBest() {
		SortResultsByBest(in)
	}
	if !rt.WithBest() {

	}

	// 1. 进行排序
	// 2. 给rank值
}
