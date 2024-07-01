package database

import (
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	//"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	//"gorm.io/gorm/logger"
)

func testNewConvenient() *convenient {
	db, err := gorm.Open(sqlite.Open("/cube/test.db"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		panic(err)
	}
	return &convenient{db: db}
}

func Test_convenient_getCubeID(t *testing.T) {
	c := testNewConvenient()
	tests := []struct {
		baseName string
	}{
		{baseName: "徐永浩"},
		{baseName: "孙大圣"},
		{baseName: "小丫鬟"},
		{baseName: "MIT-B"},
		{baseName: "熙-源~"},
		{baseName: "嘉吖"},
		{baseName: "mmmm"},
		{baseName: "徐子怡"},
	}
	for _, tt := range tests {
		t.Run(
			tt.baseName, func(t *testing.T) {
				got := c.GetCubeID(tt.baseName)
				t.Logf("%s \t %s", tt.baseName, got)
			},
		)
	}
}
