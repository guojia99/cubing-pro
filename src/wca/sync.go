// package wca

package wca

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const syncUrl = "https://www.worldcubeassociation.org/export/results/v2/sql"
const mysqlOtherSet = "?charset=utf8mb4&parseTime=True&loc=Local"
const keepDays = 1 // 只保留最近 1 天的数据（可调整）

// ==================================================================
// syncer 内部结构
// ==================================================================

type syncer struct {
	DbPath    string
	SyncPath  string
	DbURL     string
	currentDB string // e.g., "wca_20251225"

	db *gorm.DB
}

func (s *syncer) init() error {
	txtPath := filepath.Join(s.DbPath, "wca.txt")
	if data, err := os.ReadFile(txtPath); err == nil {
		s.currentDB = strings.TrimSpace(string(data))
		if !strings.HasPrefix(s.currentDB, "wca_") {
			s.currentDB = ""
		}
	}

	if err := os.MkdirAll(s.SyncPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create sync path %s: %w", s.SyncPath, err)
	}
	if err := os.MkdirAll(s.DbPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create db path %s: %w", s.DbPath, err)
	}

	return nil
}

func (s *syncer) syncFileAndSyncToDb() error {
	// Step 1: 获取远程最新数据时间
	ts, url, err := checkRemoteFileDate()
	if err != nil {
		return fmt.Errorf("check remote date failed: %w", err)
	}
	remoteDay := ts.UTC().Format("20060102")
	targetDBName := "wca_" + remoteDay
	log.Printf("Remote WCA export date: %s, target DB: %s", remoteDay, targetDBName)

	// Step 2: 如果已经使用该数据库，跳过
	if s.currentDB == targetDBName {
		log.Printf("Already using %s, skipping sync.", targetDBName)
		return nil
	}

	// Step 3: 下载 ZIP（如果不存在）
	zipPath := filepath.Join(s.SyncPath, remoteDay+".zip")
	if _, err = os.Stat(zipPath); os.IsNotExist(err) {
		log.Printf("Downloading WCA export for %s...", remoteDay)
		var err error
		zipPath, err = downloadIfNeeded(s.SyncPath, url)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
	}

	// Step 4: 解压
	extractDir := filepath.Join(s.DbPath, remoteDay)
	if _, err = os.Stat(extractDir); os.IsNotExist(err) {
		log.Printf("Extracting %s to %s...", zipPath, extractDir)
		_, err = extractZipToDb(zipPath, s.DbPath)
		if err != nil {
			return fmt.Errorf("extract failed: %w", err)
		}
	}

	sqlFile := filepath.Join(extractDir, "WCA_export.sql")
	if _, err = os.Stat(sqlFile); os.IsNotExist(err) {
		return fmt.Errorf("WCA_export.sql not found in %s", extractDir)
	}
	//  清理 SQL 文件中的 MySQL 特有注释
	log.Printf("start remove mysql version comments db file: %s", sqlFile)
	if err = cleanSQLWithSed(sqlFile); err != nil {
		return fmt.Errorf("failed to clean SQL comments: %w", err)
	}
	log.Printf("success remove mysql version comments db file: %s", sqlFile)

	// Step 5: 连接 MySQL
	dsn := s.DbURL + mysqlOtherSet
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get raw DB: %w", err)
	}
	defer sqlDB.Close()

	// Step 6: 确保目标数据库存在（先删后建，确保干净）
	if _, err = sqlDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", targetDBName)); err != nil {
		return fmt.Errorf("failed to drop existing DB %s: %w", targetDBName, err)
	}
	if _, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", targetDBName)); err != nil {
		return fmt.Errorf("failed to create DB %s: %w", targetDBName, err)
	}

	// Step 7: 导入 SQL（使用 mysql 命令行）
	if err = importSQLFileViaShell(targetDBName, sqlFile, s.DbURL); err != nil {
		// 导入失败，清理残留
		_, _ = sqlDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", targetDBName))
		return fmt.Errorf("import SQL failed: %w", err)
	}

	// Step 8: 更新索引
	log.Printf("Starting to add indexes to %s...", s.currentDB)
	if err = s.syncAddIndex(targetDBName, syncWcaDbIndex); err != nil {
		return fmt.Errorf("index creation failed: %w", err)
	}

	s.currentDB = targetDBName

	log.Printf("Successfully synced to database: %s", targetDBName)
	return nil
}

func (s *syncer) getCurrentDatabase() (*gorm.DB, string, error) {
	if s.currentDB == "" {
		return nil, "", fmt.Errorf("no current database set")
	}
	dsn := s.DbURL + s.currentDB + mysqlOtherSet
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Discard})
	s.db = db
	return db, s.currentDB, err
}

func (s *syncer) sync() error {
	if err := s.init(); err != nil {
		return fmt.Errorf("init failed: %w", err)
	}

	if err := s.syncFileAndSyncToDb(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if _, _, err := s.getCurrentDatabase(); err != nil {
		return err
	}

	// Step 9: 更新 wca.txt
	txtPath := filepath.Join(s.DbPath, "wca.txt")
	if err := os.WriteFile(txtPath, []byte(s.currentDB), 0644); err != nil {
		return fmt.Errorf("failed to update wca.txt: %w", err)
	}

	if err := s.syncStatics(); err != nil {
		log.Printf("failed to sync statics: %v", err)
	}

	if err := s.clean(); err != nil {
		log.Printf("Warning: pre-sync clean failed: %v", err)
	}

	return nil
}

//中文索引需要
//[mysqld]
//character-set-server = utf8mb4
//collation-server = utf8mb4_unicode_ci
//ft_min_word_len = 1
//ngram_token_size = 1   # 支持单字分词

const syncWcaDbIndex = `
ALTER TABLE persons ADD INDEX idx_wca_id (wca_id);

-- 全文索引 + 中文查询
ALTER TABLE persons ADD FULLTEXT(name) WITH PARSER ngram; 

-- 比赛表
CREATE INDEX idx_country_id ON competitions (country_id);
CREATE INDEX idx_year ON competitions (year);

-- 排名表
CREATE INDEX idx_person_event ON ranks_single (person_id);
CREATE INDEX idx_person_event ON ranks_average (person_id);
CREATE INDEX idx_event_world_rank ON ranks_single (event_id, world_rank);
CREATE INDEX idx_event_world_rank ON ranks_average (event_id, world_rank);
CREATE INDEX idx_event_continent_rank ON ranks_single (event_id, continent_rank);
CREATE INDEX idx_event_continent_rank ON ranks_average (event_id, continent_rank);
CREATE INDEX idx_event_country_rank ON ranks_single (event_id, country_rank);
CREATE INDEX idx_event_country_rank ON ranks_average (event_id, country_rank);

-- 成绩详情表
CREATE INDEX idx_result_id ON result_attempts (result_id);

-- 成绩表
CREATE INDEX idx_comp_id ON results (competition_id);
CREATE INDEX idx_person_id ON results (person_id);
CREATE INDEX idx_event_id ON results (event_id);
`

func (s *syncer) syncAddIndex(dbName string, indexData string) error {
	dsn := s.DbURL + mysqlOtherSet + "&multiStatements=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect for indexing: %w", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	// 切换到目标数据库
	if err = db.Exec(fmt.Sprintf("USE `%s`", dbName)).Error; err != nil {
		return fmt.Errorf("failed to use database %s: %w", dbName, err)
	}

	// 分割多条 SQL 语句（确保每条独立执行，避免 multiStatements 风险）
	statements := strings.Split(indexData, "\n")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		log.Printf("Syncing WCA index: %s", stmt)
		if err = db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("failed to execute index statement [%s]: %w", stmt, err)
		}
	}

	log.Printf("Successfully added all indexes to database: %s", dbName)
	return nil
}

// clean 清理旧文件和旧数据库
func (s *syncer) clean() error {
	now := time.Now().UTC()
	cutoff := now.AddDate(0, 0, -keepDays).Format("20060102") // 保留 >= cutoff 的数据

	// 1. 清理旧 ZIP 文件
	if entries, err := os.ReadDir(s.SyncPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".zip") {
				continue
			}
			base := strings.TrimSuffix(name, ".zip")
			if len(base) == 8 && isDigitsOnly(base) {
				if base < cutoff {
					path := filepath.Join(s.SyncPath, name)
					log.Printf("Cleaning old zip: %s", path)
					_ = os.Remove(path) // ignore error
				}
			}
		}
	}

	// 2. 清理旧解压目录
	if entries, err := os.ReadDir(s.DbPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dir := entry.Name()
			if len(dir) == 8 && isDigitsOnly(dir) {
				if dir < cutoff {
					path := filepath.Join(s.DbPath, dir)
					log.Printf("Cleaning old extract dir: %s", path)
					_ = os.RemoveAll(path)
				}
			}
		}
	}

	// 3. 清理旧数据库
	dsn := s.DbURL + mysqlOtherSet
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect for cleaning: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get raw DB for cleaning: %w", err)
	}
	defer sqlDB.Close()

	rows, err := sqlDB.Query("SHOW DATABASES")
	if err != nil {
		return fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		dbs = append(dbs, name)
	}

	for _, dbName := range dbs {
		if strings.HasPrefix(dbName, "wca_") && dbName != s.currentDB {
			datePart := strings.TrimPrefix(dbName, "wca_")
			if len(datePart) == 8 && isDigitsOnly(datePart) {
				if datePart < cutoff {
					log.Printf("Dropping old database: %s", dbName)
					sqlDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
				}
			}
		}
	}

	return nil
}

// Sync 执行完整同步流程
func (w *wca) sync() error {
	w.syncMutex.Lock()
	defer w.syncMutex.Unlock()

	s := &syncer{
		DbPath:   w.dbPath,
		SyncPath: w.syncPath,
		DbURL:    w.dbURL,
	}

	if err := s.sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	db, dbName, err := s.getCurrentDatabase()
	if err != nil {
		return fmt.Errorf("get current DB failed: %w", err)
	}

	w.db = db
	w.dbName = dbName

	if err = s.clean(); err != nil {
		log.Printf("Warning: post-sync clean failed: %v", err)
	}
	return nil
}

func (w *wca) syncLoop() {
	sy := func() {
		err := w.sync()
		if err != nil {
			log.Printf("failed to sync WCA db: %v", err)
		}
		w.updateDb()
	}

	sy()
	ticker := time.NewTicker(time.Hour * 6)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sy()
		}
	}
}

func (w *wca) updateDb() {
	txtPath := filepath.Join(w.dbPath, "wca.txt")
	wcaDbStr, err := os.ReadFile(txtPath)
	if err != nil {
		return
	}
	if len(wcaDbStr) == 0 {
		return
	}

	dsn := w.dbURL + string(wcaDbStr) + mysqlOtherSet
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Discard})
	w.db = db
}

func (w *wca) updateDbLoop() {
	ticker := time.NewTicker(time.Hour * 12)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.updateDb()
		}
	}
}
