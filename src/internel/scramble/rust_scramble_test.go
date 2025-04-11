package scramble

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func Test_scramble_rustScramble(t *testing.T) {
	cube444Scramble()

	for key, fn := range rustScrambleMp {
		t.Run(key, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				_ = fn()
			}
		})
	}

	//for key, fn := range rustCacheMp {
	//	t.Run(key, func(t *testing.T) {
	//		for i := 0; i < 10; i++ {
	//			_ = fn()
	//		}
	//	})
	//}
}

func BenchmarkRustCube444Mp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cube444Scramble()
	}
}

func Test_RustCube444Cache(t *testing.T) {

	const numTests = 100000
	const msC = 50 // 区间 ms
	timeRecords := make([]time.Duration, numTests)

	// 开始执行
	start := time.Now()
	for i := 0; i < numTests; i++ {
		begin := time.Now()
		cube444Scramble()
		timeRecords[i] = time.Since(begin)
	}
	totalTime := time.Since(start)
	fmt.Printf("Total %d execution time: %v\n", numTests, totalTime)

	// 统计耗时分布
	buckets := make(map[int][]time.Duration)
	for _, duration := range timeRecords {
		bucket := int(math.Floor(duration.Seconds() * 1000 / msC))
		buckets[bucket] = append(buckets[bucket], duration)
	}

	// 输出统计结果
	fmt.Println("Execution Time Distribution:")
	totalCount := len(timeRecords)
	for i := 0; len(buckets[i]) > 0 || len(buckets[i+1]) > 0; i++ {
		bucketData := buckets[i]
		if len(bucketData) == 0 {
			continue
		}
		total := time.Duration(0)
		for _, d := range bucketData {
			total += d
		}
		avg := total / time.Duration(len(bucketData))
		percentage := float64(len(bucketData)) / float64(totalCount) * 100
		fmt.Printf("%dms - %dms: count = %d(%.2f%%), avg = %v\n", i*msC, (i+1)*msC, len(bucketData), percentage, avg)
	}

}
