package wca

import (
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/guojia99/cubing-pro/src/wca/types"
)

// ComboResult 最终结果结构
type ComboResult struct {
	Events       []string // 组合项目列表
	SumRank      int      // 目标选手的Rank总和
	GlobalRank   int      // 在全球的真实排名 (第几名)
	TotalPlayers int      // 参与该组合排名的总人数 (分母)
}

// Task 定义一个计算任务
type Task struct {
	Mask      int
	Events    []string
	TargetSum int
}

func (s *syncer) getRanksWithCountry(country string) ([]types.RanksSingle, []types.RanksAverage) {
	var persons []types.Person
	s.db.Where("country_id = ?", country).Where("sub_id = 1").Pluck("wca_id", &persons)

	fmt.Printf("China => %d\n", len(persons))

	// 提取 ID 列表
	idList := make([]string, 0, len(persons))
	for _, person := range persons {
		// 确保 ID 不为空，避免查询出错
		if person.WcaID != "" {
			idList = append(idList, person.WcaID)
		}
	}

	var single []types.RanksSingle
	var average []types.RanksAverage

	// 3. 循环分批查询
	for i := 0; i < len(idList); i += 5000 {
		end := i + 5000
		if end > len(idList) {
			end = len(idList)
		}

		// 截取当前批次的 ID 切片
		batchIDs := idList[i:end]

		// 临时变量存储当前批次的结果
		var batchSingle []types.RanksSingle
		var batchAverage []types.RanksAverage
		s.db.Where("person_id IN ?", batchIDs).Find(&batchSingle)
		s.db.Where("person_id IN ?", batchIDs).Find(&batchAverage)

		// 4. 将结果追加到总切片中
		single = append(single, batchSingle...)
		average = append(average, batchAverage...)
	}

	return single, average
}

// FindBestGlobalCombinationsConcurrent 并发版本
func FindBestGlobalCombinationsConcurrent(allRanks []types.RanksAverage, targetID string, topN int) []ComboResult {
	// 1. 数据预处理 (同串行版)
	globalData := make(map[string]map[string]int)
	for _, r := range allRanks {
		if r.CountryRank <= 0 {
			continue
		}
		if _, ok := globalData[r.PersonID]; !ok {
			globalData[r.PersonID] = make(map[string]int)
		}
		globalData[r.PersonID][r.EventID] = r.CountryRank
	}

	targetEventsMap, ok := globalData[targetID]
	if !ok || len(targetEventsMap) == 0 {
		return []ComboResult{}
	}

	validEvents := []string{}
	for _, evt := range wcaEventsList {
		if _, exists := targetEventsMap[evt]; exists {
			validEvents = append(validEvents, evt)
		}
	}
	n := len(validEvents)

	// 预计算每个项目的默认排名：无成绩者视为 (该项目有成绩人数 + 1) 名
	// 例如 444bf 只有 150 人有成绩，则无成绩者默认 151 名
	defaultRank := make(map[string]int)
	for _, evt := range validEvents {
		count := 0
		for _, m := range globalData {
			if _, exists := m[evt]; exists {
				count++
			}
		}
		defaultRank[evt] = count + 1
	}

	// 预提取其他选手列表，避免在并发中重复操作 map
	allPersonIDs := make([]string, 0, len(globalData))
	for pid := range globalData {
		if pid != targetID {
			allPersonIDs = append(allPersonIDs, pid)
		}
	}

	fmt.Printf("启动并发计算: %d 个组合, %d 名对手, %d 个工作协程\n",
		(1<<n)-1, len(allPersonIDs), runtime.NumCPU()*2)

	// 2. 设置并发参数
	numWorkers := runtime.NumCPU() * 2 // 使用 CPU 核心数
	taskChan := make(chan Task, 1000)  // 缓冲通道，防止发送阻塞
	resultChan := make(chan ComboResult, 1000)

	var wg sync.WaitGroup

	// 3. 启动 Worker Pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				// --- 核心计算逻辑 (每个 Worker 独立执行) ---
				// 参与者：至少有一项组合内成绩的人；无成绩的项目用 defaultRank
				betterCount := 0
				totalQualified := 0

				for _, otherID := range allPersonIDs {
					otherRanks := globalData[otherID]
					hasAtLeastOne := false
					otherSum := 0

					for _, evt := range task.Events {
						rank, exists := otherRanks[evt]
						if exists {
							hasAtLeastOne = true
							otherSum += rank
						} else {
							otherSum += defaultRank[evt] // 无成绩默认排名
						}
					}

					if hasAtLeastOne {
						totalQualified++
						if otherSum < task.TargetSum {
							betterCount++
						}
					}
				}

				// 排名规则：总分相同则排名相同，下一名顺延 (如 1, 2, 2, 4)
				// betterCount = 严格优于目标的人数，故 rank = betterCount + 1
				resultChan <- ComboResult{
					Events:       task.Events,
					SumRank:      task.TargetSum,
					GlobalRank:   betterCount + 1,
					TotalPlayers: totalQualified + 1, // +1 包含自己
				}
			}
		}()
	}

	// 4. 发送任务 (在主协程中)
	totalCombos := 1 << n
	go func() {
		for mask := 1; mask < totalCombos; mask++ {
			currentEvents := []string{}
			targetSum := 0
			for i := 0; i < n; i++ {
				if (mask & (1 << i)) != 0 {
					evt := validEvents[i]
					currentEvents = append(currentEvents, evt)
					targetSum += targetEventsMap[evt]
				}
			}
			taskChan <- Task{
				Mask:      mask,
				Events:    currentEvents,
				TargetSum: targetSum,
			}
		}
		close(taskChan) // 任务发送完毕，关闭通道，通知 Workers 结束
	}()

	// 5. 等待所有 Worker 完成并关闭结果通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 6. 收集结果
	results := make([]ComboResult, 0, totalCombos)
	for res := range resultChan {
		results = append(results, res)
	}

	// 7. 排序 (排序本身很快，不需要并发)
	sort.Slice(results, func(i, j int) bool {
		if results[i].GlobalRank != results[j].GlobalRank {
			return results[i].GlobalRank < results[j].GlobalRank
		}
		if results[i].SumRank != results[j].SumRank {
			return results[i].SumRank < results[j].SumRank
		}
		return len(results[i].Events) < len(results[j].Events)
	})

	if len(results) > topN {
		results = results[:topN]
	}

	return results
}
