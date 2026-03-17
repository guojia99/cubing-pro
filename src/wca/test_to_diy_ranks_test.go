package wca

import (
	"fmt"
	"testing"
)

func TestFindBestGlobalCombinations(t *testing.T) {
	s := &syncer{
		DbURL:     "root@tcp(127.0.0.1:33036)/",
		currentDB: curTestDb,
	}

	_, _, err := s.getCurrentDatabase()
	if err != nil {
		t.Fatal(err)
	}

	_, avg := s.getRanksWithCountry("China")
	targetID := "2017HUAN77"
	top5 := FindBestGlobalCombinationsConcurrent(avg, targetID, 5)
	fmt.Printf("\n=== %s 最佳自定义组合 Top 5 ===\n", targetID)
	for i, res := range top5 {
		fmt.Printf("No.%d\n", i+1)
		fmt.Printf("  组合: %v\n", res.Events)
		fmt.Printf("  个人总分: %d\n", res.SumRank)
		fmt.Printf("  全国排名: #%d / %d 人\n", res.GlobalRank, res.TotalPlayers)
		fmt.Println("-------------------------")
	}
}
