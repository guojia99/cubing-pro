package wca

import (
	"sync"
	"time"

	"github.com/guojia99/cubing-pro/src/wca/types"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type WCA interface {
	// ExportToSqlite 导出
	ExportToSqlite(sqlitePath string) error // 耗时较长，需要较多内存来存放数据，仅实验使用
	ExportToTable(filePath string) error

	// wca查询类
	SearchPlayers(name string) []types.Person

	GetPersonInfo(wcaId string) (types.PersonInfo, error)
	GetPersonResult(wcaId string) ([]types.Result, error)
	GetCompetition(compId string) (types.Competition, error)
	GetPersonCompetition(wcaId string) ([]types.Competition, error)

	// 统计
	GetPersonRankTimer(wcaId string) ([]types.StaticWithTimerRank, error)
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

func NewWCA(
	mysqlUrl string,
	dbPath string,
	syncPath string,
	enableSync bool,
) (WCA, error) {
	var err error
	w := &wca{
		dbPath:   dbPath,
		syncPath: syncPath,
		dbURL:    mysqlUrl,
		cache:    cache.New(5*time.Minute, 10*time.Minute),
	}
	w.updateDb()
	if enableSync {
		// 初始化后必须立即同步数据库
		if err = w.sync(); err != nil {
			return nil, err
		}
		go w.syncLoop()
	} else {
		go w.updateDbLoop()
	}

	return w, nil
}
