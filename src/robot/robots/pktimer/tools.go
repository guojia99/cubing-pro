package pktimer

import (
	"sort"
)

const eps = 0.015 // 1.5%

type cachePlayer struct {
	UserName  string
	CurResult float64
}

// GroupPlayersBySimilarScore 将成绩在1%范围内的玩家聚合在一起
func groupPlayersBySimilarScore(players []cachePlayer) map[int][]cachePlayer {
	result := make(map[int][]cachePlayer)
	if len(players) == 0 {
		return result
	}

	// 按成绩排序
	sort.Slice(players, func(i, j int) bool {
		return players[i].CurResult < players[j].CurResult
	})

	var groups [][]cachePlayer

	// 贪心分组：每组以第一个成员为基准
	for _, p := range players {
		if len(groups) == 0 {
			groups = append(groups, []cachePlayer{p})
			continue
		}

		// 获取当前组的基准成绩（第一个成员）
		base := groups[len(groups)-1][0].CurResult
		if isWithinOnePercent(p.CurResult, base, eps) {
			groups[len(groups)-1] = append(groups[len(groups)-1], p)
		} else {
			// 新建一组
			groups = append(groups, []cachePlayer{p})
		}
	}

	// 转成 map[int][]cachePlayer，key 为组索引
	for i, group := range groups {
		result[i] = group
	}

	return result
}

// isWithinOnePercent 判断 a 和 b 是否在 1% 相对误差范围内
func isWithinOnePercent(a, b, eps float64) bool {
	if a == b {
		return true
	}
	return getDiffPercent(a, b) <= eps
}

func getDiffPercent(a, b float64) float64 {
	if a <= b {
		a, b = b, a
	}
	maxVal := a
	if b > a {
		maxVal = b
	}
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff / maxVal
}
