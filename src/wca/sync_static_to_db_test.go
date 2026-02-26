package wca

import (
	"fmt"
	"testing"
	"time"

	"github.com/guojia99/cubing-pro/src/test_tool"
	"github.com/guojia99/cubing-pro/src/wca/types"
	jsoniter "github.com/json-iterator/go"
)

const curTestDb = "wca_20251228"

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
	_ = NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path",
		true)

}

func Test_syncer_setStaticPersonRankWithTimer(t *testing.T) {
	// 启动内存监控
	mon := test_tool.NewMemMonitor()
	mon.Start(time.Second * 3) // 每秒采样一次
	defer mon.Stop()           // 确保测试结束时停止

	s := &syncer{
		DbURL:     "root@tcp(127.0.0.1:33306)/",
		currentDB: curTestDb,
	}
	_, _, err := s.getCurrentDatabase()
	if err != nil {
		t.Fatal(err)
	}
	if err = s.setStaticPersonRankWithTimers(); err != nil {
		t.Fatal(err)
	}
}

func Test_syncer_getResultMapWithEvent(t *testing.T) {
	// 启动内存监控
	mon := test_tool.NewMemMonitor()
	mon.Start(time.Second * 3) // 每秒采样一次
	defer mon.Stop()           // 确保测试结束时停止

	s := &syncer{
		DbURL:     "root@tcp(127.0.0.1:33306)/",
		currentDB: curTestDb,
	}

	_, _, err := s.getCurrentDatabase()
	if err != nil {
		t.Fatal(err)
	}

	data := s.getResultMapWithEvent("333bf")

	d, _ := jsoniter.MarshalIndent(data["PleaseBeQuietHefei2025"], "", "    ")
	fmt.Println(string(d))

}

func Test_getCurPersonsRankTimerSnapshots(t *testing.T) {
	s := &syncer{
		DbURL:     "root@tcp(127.0.0.1:33306)/",
		currentDB: curTestDb,
	}

	_, _, err := s.getCurrentDatabase()
	if err != nil {
		t.Fatal(err)
	}

	curAllPersonValue := make(map[string]*curPersonValue)
	results := s.getResultMapWithEvent("333bf")

	var allResults []types.Result

	for _, l := range results {
		allResults = append(allResults, l...)
	}
	s.updateStaticCurAllPersonValue(curAllPersonValue, allResults, s.countryMap())
	data := curAllPersonValue["2017WANY29"]
	d, _ := jsoniter.MarshalIndent(data, "", "    ")
	fmt.Println(string(d))

	out := s.getCurPersonsRankTimerSnapshots("333bf", time.Now(), curAllPersonValue)
	for _, v := range out {
		if v.WcaID == "2017WANY29" {
			d, _ = jsoniter.MarshalIndent(v, "", "    ")
			fmt.Println(string(d))
		}
	}

}
