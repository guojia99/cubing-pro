package wca

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/guojia99/cubing-pro/src/wca/types"
	utils_tool "github.com/guojia99/cubing-pro/src/wca/utils"
)

func (w *wca) getBestWrN(event string, wrN int) []string {
	var withAvg bool = true
	switch event {
	case "333bf", "444bf", "555bf":
		withAvg = false
	}

	var out []string
	if wrN == 0 {
		wrN = 500
	}
	if withAvg {
		var bestWrNAvgRanks []types.RanksAverage
		w.db.Where("event_id = ?", event).
			Order("world_rank ASC").
			Limit(wrN).
			Find(&bestWrNAvgRanks)

		for _, b := range bestWrNAvgRanks {
			out = append(out, b.PersonID)
		}
	} else {
		var bestWrNSingleRanks []types.RanksSingle
		w.db.Where("event_id = ?", event).
			Order("world_rank ASC").
			Limit(wrN).
			Find(&bestWrNSingleRanks)
		for _, b := range bestWrNSingleRanks {
			out = append(out, b.PersonID)
		}
	}
	return out
}

func getCompetitionIDYear(id string) (int, error) {
	if len(id) < 4 {
		return 0, fmt.Errorf("string length less than 4")
	}

	last4 := id[len(id)-4:]
	n, err := strconv.Atoi(last4)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *wca) getPersonsResultWithEvent(personID []string, event string) map[string][]types.Result {
	var results []types.Result
	w.db.Where("event_id = ?", event).Where("person_id IN (?)", personID).Find(&results)

	sort.Slice(results, func(i, j int) bool {
		return results[i].ID > results[j].ID
	})

	// 每个人只取最近10条有效成绩
	var useResults = make([]types.Result, 0)
	var out = make(map[string][]types.Result)
	for _, r := range results {
		if r.Best <= 0 {
			continue
		}
		if year, err := getCompetitionIDYear(r.CompetitionID); err != nil || year <= 2018 {
			continue
		}
		list := out[r.PersonID]
		if len(list) >= 10 {
			continue
		}
		out[r.PersonID] = append(list, r)
		useResults = append(useResults, r)
	}

	// Attempts
	attemptMap := w.getResultAttemptMap(useResults)
	for k, list := range out {
		for idx, r := range list {
			list[idx].Attempts = attemptMap[r.ID]
		}
		out[k] = list
	}

	return out
}

// resultProportionEstimationData 根据多名选手在近期成绩（Result + Attempts）上的统计，估计各 WCA 项目之间的「合理比例曲线」。
//
// ## 数据与符号
// - 输入 data：personID -> (eventID -> 该选手该项目下至多 10 条 Result)，每条 Result 已附带 Attempts（百分之一秒；DNF 等为 ≤0，全部丢弃）。
// - events：须与 ResultProportionEstimationMap 中该类型顺序一致；events[0] 为**锚点项目**（用户侧「已知一项求其他项」时以该项为自变量）。
// - 时间量纲：内部一律用 WCA 的**百分之一秒 (centiseconds)** 做比例，输出曲线采样中另给出秒 (sec) 便于展示。
//
// ## 算法依据
//  1. **个体表征（每位选手一条向量）**
//     同一选手在不同轮次、不同赛事上的单次成绩来自 Attempts。要求「最近的数据」已在采集层通过每人每项目取最近若干条 Result 满足。
//     对每位选手、每个项目，收集该选手在该项目所有 Result 上、全部 Attempts 中 **>0** 的值，取**中位数**作为该项目上的代表时间。
//     中位数对「一轮里偶尔崩掉的一次」不敏感，比单次 Best 或简单平均更稳，适合描述「该选手当前水平带」。
//  2. **比例模型（项目间关系）**
//     对选手 p、项目 j，记中位数时间为 T_pj，锚点项目 a=events[0]。定义比例 r_pj = T_pj / T_pa（j=a 时为 1）。
//     在「水平相近的顶尖选手」中，r 的分布应较集中；取群体在段内的 **r 的中位数** 作为该段上 j 相对 a 的标度关系。
//     这等价于在 log 域用「加性」差异近似乘性比例，对魔方大魔方等近似乘性增长的时间结构较自然。
//  3. **分段（相近成绩归并）**
//     全体选手按 T_pa（锚点中位数）升序排列。将选手按**等人数分箱**（分箱数约 sqrt(N)，并限制在 [1, 30]），使每一档内选手锚点水平接近，从而「相近成绩拟合到一起」。
//     每箱内对除锚点外的每个项目，计算 {r_pj} 的**中位数** 得到该段的 Ratio，并记录箱内锚点时间与人数。
//     样本过少（全程不足约 5 人）时退化为**单一段**，仅用全局比例，避免过拟合噪声。
//  4. **光滑曲线（分段间的插值）**
//     每个分段有中位锚点区间 [AnchorMin, AnchorMax]，取段内锚点的最小、最大作为边界（整数厘秒）。
//     对每个非锚点项目，在各段中心锚点（段内锚点中位数）处有一条阶梯式的「段比例」；为得到连续参考曲线，在**段中心**之间对**比例**做一维**线性插值**；
//     低于最低段中心或高于最高段中心时，**常数外推**为最近一段的比例（避免无依据的线性外推爆炸）。
//  5. **预测用法**
//     给定锚点成绩 t_a（厘秒），先对各项目 j≠a 得到插值后的 ratio_j(t_a)，再估计 t_j = t_a * ratio_j。
//     示例：bigcube 锚点为 444，若 t_a=1900（19.00s），则 555≈1900*ratio_555 厘秒，换算为秒即用户直觉数值。
//
// ## 局限
// - 不同选手「444 快但 666 相对慢」的个体结构差异被群体统计平滑；输出是: 参考中位结构，非个体预测。
func resultProportionEstimationData(events []string, data map[string]map[string][]types.Result) (out types.ResultProportionEstimationResult, err error) {
	if len(events) < 2 {
		return out, errors.New("need at least 2 events")
	}
	anchor := events[0]

	rows := buildProportionPersonRows(events, data)
	if len(rows) == 0 {
		return out, errors.New("no person with valid attempts on all events")
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Anchor < rows[j].Anchor
	})

	globalRatio := medianRatiosOverRows(rows, events, anchor)
	segments := buildProportionSegments(rows, events, anchor)

	out.Events = append([]string(nil), events...)
	out.GlobalRatio = globalRatio
	out.SampleCount = len(rows)
	out.Segments = segments

	minA, maxA := rows[0].Anchor, rows[len(rows)-1].Anchor
	out.CurveSamples = sampleProportionCurve(events, anchor, segments, minA, maxA, globalRatio)

	return out, nil
}

// ResultProportionEstimation
// 1. 取各个项目wrN （N暂定为500），取交集得出四个项目都有的选手。
// 2. 取各个选手最近10轮成绩（去掉异常数据，如DNF，如偏离wrN太远的成绩，避免突然进步哥）
// 3. 删除时间太久远的成绩，只保留19年以后的成绩。
// 4. 依据以上数据集，静态拟合各个节点相近的成绩。
// 5. 依据比例计算出各个项目在不同节点下推测的成绩。
// ## 算法
//
// - **个体表征**
//   - 同一选手在不同轮次、不同赛事上的单次成绩来自 Attempts。要求「最近的数据」已在采集层通过每人每项目取最近若干条 Result 满足。
//   - 对每位选手、每个项目，收集该选手在该项目所有 Result 上、全部 Attempts 中 **>0** 的值，取**中位数**作为该项目上的代表时间。
//
// - **比例模型**
//   - 对选手 p、项目 j，记中位数时间为 T_pj，锚点项目 a=events[0]。定义比例 r_pj = T_pj / T_pa（j=a 时为 1）。
//   - 在「水平相近的顶尖选手」中，r 的分布应较集中；取群体在段内的 **r 的中位数** 作为该段上 j 相对 a 的标度关系。
//   - 等价于在 log 域用「加性」差异近似乘性比例，对魔方大魔方等近似乘性增长的时间结构较自然。
//
// - **分段（相近成绩归并）**
//   - 全体选手按 T_pa（锚点中位数）升序排列。将选手按**等人数分箱**（分箱数约 sqrt(N)，并限制在 [1, 30]），使每一档内选手锚点水平接近，从而「相近成绩拟合到一起」。
//   - 每箱内对除锚点外的每个项目，计算 {r_pj} 的**中位数** 得到该段的 Ratio，并记录箱内锚点时间与人数。
//   - 样本过少（全程不足约 5 人）时退化为**单一段**，仅用全局比例，避免过拟合噪声。
//
// - **光滑曲线（分段间的插值）**
//   - 每个分段有中位锚点区间 [AnchorMin, AnchorMax]，取段内锚点的最小、最大作为边界（整数厘秒）。
//   - 对每个非锚点项目，在各段中心锚点（段内锚点中位数）处有一条阶梯式的「段比例」；为得到连续参考曲线，在**段中心**之间对**比例**做一维**线性插值**；
//   - 低于最低段中心或高于最高段中心时，**常数外推**为最近一段的比例（避免无依据的线性外推爆炸）。
//
// - **预测用法**
//   - 给定锚点成绩 t_a（厘秒），先对各项目 j≠a 得到插值后的 ratio_j(t_a)，再估计 t_j = t_a * ratio_j。
//   - 示例：bigcube 锚点为 444，若 t_a=1900（19.00s），则 555≈1900*ratio_555 厘秒，换算为秒即用户直觉数值。
func (w *wca) ResultProportionEstimation(estimationType types.ResultProportionEstimationType, WrN int) (out types.ResultProportionEstimationResult, err error) {

	events, typeOK := types.ResultProportionEstimationMap[estimationType]
	if !typeOK {
		return types.ResultProportionEstimationResult{}, errors.New("invalid estimationType")
	}

	// 获取交集的人
	var cachePersons [][]string
	for _, event := range events {
		cachePersons = append(cachePersons, w.getBestWrN(event, WrN))
	}
	persons := utils_tool.IntersectAll(cachePersons) // 交集

	// 获取所有人的成绩列表
	// map[personID]map[eventID][]types.Result
	var baseData = make(map[string]map[string][]types.Result)
	for _, event := range events {
		results := w.getPersonsResultWithEvent(persons, event)
		for personID, result := range results {
			if _, ok := baseData[personID]; !ok {
				baseData[personID] = make(map[string][]types.Result)
			}
			baseData[personID][event] = result
		}
	}
	// 数据采集完成, 执行分析
	out, err = resultProportionEstimationData(events, baseData)
	if err != nil {
		return
	}
	out.Persons = persons
	return out, err
}

// proportionPersonRow 表示一名选手在各项目上的代表时间（本人近期全部有效 Attempts 的中位数，厘秒）。
type proportionPersonRow struct {
	PersonID string
	Medians  map[string]float64
	Anchor   float64
}

func validAttemptsFromResults(results []types.Result) []int64 {
	var out []int64
	for _, r := range results {
		for _, a := range r.Attempts {
			if a > 0 {
				out = append(out, a)
			}
		}
	}
	return out
}

func medianInt64(xs []int64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]int64(nil), xs...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	n := len(cp)
	if n%2 == 1 {
		return float64(cp[n/2])
	}
	return float64(cp[n/2-1]+cp[n/2]) / 2
}

func medianFloat64(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	sort.Float64s(cp)
	n := len(cp)
	if n%2 == 1 {
		return cp[n/2]
	}
	return (cp[n/2-1] + cp[n/2]) / 2
}

func buildProportionPersonRows(events []string, data map[string]map[string][]types.Result) []proportionPersonRow {
	anchor := events[0]
	var rows []proportionPersonRow
	for pid, evMap := range data {
		medians := make(map[string]float64)
		ok := true
		for _, e := range events {
			rs := evMap[e]
			att := validAttemptsFromResults(rs)
			if len(att) == 0 {
				ok = false
				break
			}
			medians[e] = medianInt64(att)
		}
		if !ok {
			continue
		}
		rows = append(rows, proportionPersonRow{
			PersonID: pid,
			Medians:  medians,
			Anchor:   medians[anchor],
		})
	}
	return rows
}

func medianRatiosOverRows(rows []proportionPersonRow, events []string, anchor string) map[string]float64 {
	out := make(map[string]float64)
	for _, e := range events {
		if e == anchor {
			continue
		}
		var ratios []float64
		for _, r := range rows {
			if r.Anchor <= 0 {
				continue
			}
			ratios = append(ratios, r.Medians[e]/r.Anchor)
		}
		out[e] = medianFloat64(ratios)
	}
	return out
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func buildProportionSegments(rows []proportionPersonRow, events []string, anchor string) []types.ProportionEstimationSegment {
	n := len(rows)
	if n == 0 {
		return nil
	}
	if n < 5 {
		return []types.ProportionEstimationSegment{segmentFromSubRows(rows, events, anchor)}
	}

	numBins := int(math.Round(math.Sqrt(float64(n))))
	numBins = clampInt(numBins, 1, 30)
	if numBins > n {
		numBins = n
	}

	var segs []types.ProportionEstimationSegment
	for b := 0; b < numBins; b++ {
		start := b * n / numBins
		end := (b + 1) * n / numBins
		if start >= end {
			continue
		}
		sub := rows[start:end]
		segs = append(segs, segmentFromSubRows(sub, events, anchor))
	}
	if len(segs) == 0 {
		return []types.ProportionEstimationSegment{segmentFromSubRows(rows, events, anchor)}
	}
	return segs
}

func segmentFromSubRows(sub []proportionPersonRow, events []string, anchor string) types.ProportionEstimationSegment {
	var anchors []float64
	for _, r := range sub {
		anchors = append(anchors, r.Anchor)
	}
	sort.Float64s(anchors)
	amin := int(math.Round(anchors[0]))
	amax := int(math.Round(anchors[len(anchors)-1]))
	if amin == amax {
		if amin > 1 {
			amin--
		}
		amax++
	}

	ratio := make(map[string]float64)
	for _, e := range events {
		if e == anchor {
			continue
		}
		var ratios []float64
		for _, r := range sub {
			if r.Anchor <= 0 {
				continue
			}
			ratios = append(ratios, r.Medians[e]/r.Anchor)
		}
		ratio[e] = medianFloat64(ratios)
	}

	return types.ProportionEstimationSegment{
		AnchorMin: amin,
		AnchorMax: amax,
		NPersons:  len(sub),
		Ratio:     ratio,
	}
}

func interpolateRatioAt(anchorVal float64, segs []types.ProportionEstimationSegment, event string, globalRatio float64) float64 {
	if len(segs) == 0 {
		return globalRatio
	}
	type crPair struct {
		c, r float64
	}
	ps := make([]crPair, 0, len(segs))
	for _, s := range segs {
		c := (float64(s.AnchorMin) + float64(s.AnchorMax)) / 2
		r := 0.0
		if s.Ratio != nil {
			r = s.Ratio[event]
		}
		if r <= 0 {
			r = globalRatio
		}
		ps = append(ps, crPair{c, r})
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].c < ps[j].c })

	if len(ps) == 1 {
		return ps[0].r
	}
	if anchorVal <= ps[0].c {
		return ps[0].r
	}
	last := ps[len(ps)-1]
	if anchorVal >= last.c {
		return last.r
	}
	for i := 0; i < len(ps)-1; i++ {
		if anchorVal >= ps[i].c && anchorVal <= ps[i+1].c {
			d := ps[i+1].c - ps[i].c
			if d <= 0 {
				return ps[i].r
			}
			w := (anchorVal - ps[i].c) / d
			return ps[i].r*(1-w) + ps[i+1].r*w
		}
	}
	return last.r
}

func sampleProportionCurve(events []string, anchor string, segments []types.ProportionEstimationSegment, minA, maxA float64, global map[string]float64) []types.ProportionCurveSample {
	const nSamples = 50
	segsSorted := append([]types.ProportionEstimationSegment(nil), segments...)
	sort.Slice(segsSorted, func(i, j int) bool {
		mi := (float64(segsSorted[i].AnchorMin) + float64(segsSorted[i].AnchorMax)) / 2
		mj := (float64(segsSorted[j].AnchorMin) + float64(segsSorted[j].AnchorMax)) / 2
		return mi < mj
	})

	den := float64(nSamples - 1)
	if den < 1 {
		den = 1
	}

	if minA >= maxA {
		maxA = minA + 1
	}

	out := make([]types.ProportionCurveSample, 0, nSamples)
	for i := 0; i < nSamples; i++ {
		t := float64(i) / den
		a := minA*(1-t) + maxA*t
		est := make(map[string]float64)
		for _, e := range events {
			if e == anchor {
				est[e] = a / 100
				continue
			}
			gr := global[e]
			if gr <= 0 {
				gr = 1
			}
			r := interpolateRatioAt(a, segsSorted, e, gr)
			est[e] = a * r / 100
		}
		out = append(out, types.ProportionCurveSample{
			AnchorSec: a / 100,
			Estimates: est,
		})
	}
	return out
}
