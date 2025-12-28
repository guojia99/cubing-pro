package wca

import (
	"fmt"
	"testing"
	"time"

	"github.com/guojia99/cubing-pro/src/test_tool"
)

func Test_getEndCompsTimer(t *testing.T) {
	now := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	fmt.Println(getEndCompsTimer(now)) // 输出: 2023-01-07

	now = time.Date(2023, 12, 31, 10, 0, 0, 0, time.UTC)
	fmt.Println(getEndCompsTimer(now)) // 输出: 2023-12-31

	now = time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
	fmt.Println(getEndCompsTimer(now)) // 2023-06-17（因为1月1日是周日，第24周结束于6月17日）

	now = time.Now()
	fmt.Println(getEndCompsTimer(now)) // 2023-06-17（因为1月1日是周日，第24周结束于6月17日）
}

func Test_wca_getStaticPersonRankWithTimer(t *testing.T) {

	// 启动内存监控
	mon := test_tool.NewMemMonitor()
	mon.Start(time.Second) // 每秒采样一次
	defer mon.Stop()       // 确保测试结束时停止

	_, err := NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path")
	if err != nil {
		t.Fatal(err)
	}

}
