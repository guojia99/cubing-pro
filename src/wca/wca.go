package wca

import (
	"fmt"
	"sync"
	"time"

	"github.com/guojia99/cubing-pro/src/robot/qq_bot/Better-Bot-Go/log"
	"github.com/guojia99/cubing-pro/src/wca/types"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

// BaseWCA 基础统计功能
type BaseWCA interface {
	// ExportToSqlite 导出
	ExportToSqlite(sqlitePath string) error // 耗时较长，需要较多内存来存放数据，仅实验使用
	ExportToTable(filePath string) error

	// wca查询类
	SearchPlayers(name string) []types.Person
	CountryList() []types.Country

	GetPersonInfo(wcaId string) (types.PersonInfo, error)
	GetPersonResult(wcaId string) ([]types.Result, error)
	GetCompetition(compId string) (types.Competition, error)
	GetPersonCompetition(wcaId string) ([]types.Competition, error)

	// 大满贯列表
	GetGrandSlam() []types.AllEventChampionshipsPodium

	// 统计
	GetPersonRankTimer(wcaId string) ([]types.StaticWithTimerRank, error)
	GetEventRankWithTimer(eventId, country string, year int, isAvg bool, page, size int) ([]types.StaticWithTimerRank, int64, error)

	// GetEventRankWithFullNow 给出现在成绩的排序
	GetEventRankWithFullNow(eventId, country string, isAvg bool, page, size int) ([]types.Result, int64, error)
	// GetEventRankWithOnlyYear 只计算当年成绩的排序
	GetEventRankWithOnlyYear(eventId, countryID string, year int, month int, isAvg bool, page, size int) ([]types.Result, int64, error)
	// GetEventSuccessRateResult 成功率
	GetEventSuccessRateResult(eventId, country string, minAttempted, page, size int) ([]types.StaticSuccessRateResult, int64, error)

	// GetPersonBestRanks 获取选手最佳成绩排行
	GetPersonBestRanks(wcaID string) (types.PersonBestRanks, error)

	// GetAllEventsAchievement 全项目达成check
	GetAllEventsAchievement(lackNum int, country string, page int, size int) ([]types.AllEventAvgPersonResults, int64, error)

	// GetRankWithEvents 根据项目列表进行排序
	GetRankWithEvents(events []string, country string, avg bool, page int, size int) (out []types.RankWithEventsStatic, count int64, err error)

	// GetCountryBestWithEventGroupRank 获取选手最佳项目排列
	GetCountryBestWithEventGroupRank(wcaId string, avg bool, useWorld bool) (out []types.RankWithEventsGrouptatic, err error)

	// GetWithCompYearPersonRank 获取相同年限参赛选手最佳成绩排行
	GetWithCompYearPersonRank(year int, country string, eventID string, avg bool, page int, size int) (out []types.RankWithPersonCompStartYear, count int64, err error)

	GetNotPodiumSor(events []string, country string, bestMisser int, avg bool, page int, size int) (out []types.RankWithEventsStatic, count int64, err error)
}

// ExtendWCA 拓展功能
type ExtendWCA interface {
	ResultProportionEstimation(estimationType types.ResultProportionEstimationType, WrN int) (types.ResultProportionEstimationResult, error)
}

type WCA interface {
	BaseWCA

	ExtendWCA

	SyncStatic() error
}

type wca struct {
	db     *gorm.DB
	dbName string

	syncMutex sync.Mutex
	// 目录结构
	// ---->
	//    /wca
	//    /wca/zipPath
	//    /wca/syncDb
	dbPath   string
	syncPath string
	dbURL    string

	cache *cache.Cache
}

func (w *wca) SyncStatic() error {
	s := &syncer{
		DbPath:   w.dbPath,
		SyncPath: w.syncPath,
		DbURL:    w.dbURL,
	}

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	return nil
}

func NewWCA(
	mysqlUrl string,
	dbPath string,
	syncPath string,
	enableSync bool,
) WCA {
	w := &wca{
		dbPath:   dbPath,
		syncPath: syncPath,
		dbURL:    mysqlUrl,
		cache:    cache.New(5*time.Minute, 10*time.Minute),
	}
	w.updateDb()
	if w.db == nil {
		log.Errorf("sync wca db is failed")
	}

	log.Infof("sync wca db and start loop: %+v", enableSync)
	if enableSync {
		go w.syncLoop()
	} else {
		go w.updateDbLoop()
	}

	return w
}
