package wca

import (
	"fmt"
	"os"
	"reflect"

	"github.com/guojia99/cubing-pro/src/wca/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const batchSize = 1024 // 每批处理行数，可调

// ExportToSqlite 将 MySQL 中的 WCA 数据分页导出到 SQLite 文件，并实时显示进度
func (w *wca) ExportToSqlite(sqlitePath string) error {
	_ = os.Remove(sqlitePath)

	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return fmt.Errorf("open sqlite db: %w", err)
	}

	// 先创建所有表结构
	models := []interface{}{
		&types.Championship{},
		&types.Competition{},
		&types.Continent{},
		&types.Country{},
		&types.EligibleCountryISO2ForChampionship{},
		&types.Event{},
		&types.Format{},
		&types.Person{},
		&types.RanksAverage{},
		&types.RanksSingle{},
		&types.ResultAttempt{},
		&types.Result{},
		&types.RoundType{},
		&types.SchemaMigration{},
		&types.Scramble{},
	}
	for _, m := range models {
		if err := sqliteDB.AutoMigrate(m); err != nil {
			return fmt.Errorf("auto migrate: %w", err)
		}
	}

	// 按顺序迁移每张表（带进度）
	tables := []struct {
		name string
		new  func() interface{} // 用于获取空实例
	}{
		{"championships", func() interface{} { return &types.Championship{} }},
		{"competitions", func() interface{} { return &types.Competition{} }},
		{"continents", func() interface{} { return &types.Continent{} }},
		{"countries", func() interface{} { return &types.Country{} }},
		{"eligible_country_iso2s_for_championship", func() interface{} { return &types.EligibleCountryISO2ForChampionship{} }},
		{"events", func() interface{} { return &types.Event{} }},
		{"formats", func() interface{} { return &types.Format{} }},
		{"persons", func() interface{} { return &types.Person{} }},
		{"ranks_average", func() interface{} { return &types.RanksAverage{} }},
		{"ranks_single", func() interface{} { return &types.RanksSingle{} }},
		{"result_attempts", func() interface{} { return &types.ResultAttempt{} }},
		{"results", func() interface{} { return &types.Result{} }},
		{"round_types", func() interface{} { return &types.RoundType{} }},
		{"schema_migrations", func() interface{} { return &types.SchemaMigration{} }},
		{"scrambles", func() interface{} { return &types.Scramble{} }},
	}

	for _, tbl := range tables {
		fmt.Printf("📦 Migrating %s...\n", tbl.name)
		if err = w.migrateTable(sqliteDB, tbl.name, tbl.new()); err != nil {
			return fmt.Errorf("migrate table %s: %w", tbl.name, err)
		}
	}

	fmt.Println("✅ Export to SQLite completed!")
	return nil
}

// migrateTable 泛型辅助函数：分页读取源表并写入 SQLite，实时输出进度
// migrateTable 全量读取 MySQL 表（一次查询），分批插入 SQLite
func (w *wca) migrateTable(db *gorm.DB, tableName string, emptyModel interface{}) error {
	// 1. 获取模型类型（必须是指针）
	modelType := reflect.TypeOf(emptyModel)
	if modelType.Kind() != reflect.Ptr {
		return fmt.Errorf("emptyModel must be a pointer to a struct")
	}
	elemType := modelType.Elem()

	// 2. 创建切片类型：[]T
	sliceType := reflect.SliceOf(elemType)
	slicePtr := reflect.New(sliceType) // []*T

	// 3. 全量查询 MySQL
	fmt.Printf("  ➤ Reading all records from MySQL table '%s'...\n", tableName)
	if err := w.db.Find(slicePtr.Interface()).Error; err != nil {
		return fmt.Errorf("failed to read %s: %w", tableName, err)
	}

	// 4. 获取实际数据 slice
	sliceValue := slicePtr.Elem() // []T
	total := sliceValue.Len()

	if total == 0 {
		fmt.Printf("  ➤ [%s] 0 / 0 (100.0%%) - no data\n", tableName)
		return nil
	}

	fmt.Printf("  ➤ Loaded %d records. Inserting into SQLite in batches...\n", total)

	// 5. 分批插入 SQLite
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		// 构造子 slice: records[i:end]
		batchSlice := sliceValue.Slice(i, end)
		batchPtr := reflect.New(batchSlice.Type()) // 创建新 slice 指针
		batchPtr.Elem().Set(batchSlice)

		// 批量插入
		if err := db.CreateInBatches(batchPtr.Interface(), batchSize).Error; err != nil {
			return fmt.Errorf("insert batch [%d-%d) into %s: %w", i, end, tableName, err)
		}

		// 进度
		done := end
		percent := float64(done) * 100 / float64(total)
		fmt.Printf("➤ [%s] %d / %d (%.1f%%)\n", tableName, done, total, percent)
	}

	fmt.Printf("➤ [%s] %d / %d (100.0%%)\n", tableName, total, total)
	return nil
}
