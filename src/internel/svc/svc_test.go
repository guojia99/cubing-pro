package svc

import (
	"testing"

	"github.com/guojia99/cubing-pro/src/configs"
)

func TestNewDBRejectsUnsupportedDrivers(t *testing.T) {
	for _, driver := range []string{"", "sqlite", "postgres"} {
		t.Run(driver, func(t *testing.T) {
			var cfg configs.GlobalConfig
			cfg.DB.Driver = driver
			db, err := newDB(cfg)
			if err == nil || db != nil {
				t.Fatalf("driver %q: got database %v, error %v; want unsupported driver error", driver, db, err)
			}
		})
	}
}
