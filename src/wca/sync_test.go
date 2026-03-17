package wca

import (
	"testing"
)

func Test_wca_Sync(t *testing.T) {
	s := syncer{
		DbPath:   "/Users/guojia/data/cubingPro/wca_db",
		SyncPath: "/Users/guojia/data/cubingPro/wca_db/sync_path",
		DbURL:    "root@tcp(127.0.0.1:33036)/",
	}
	if err := s.init(); err != nil {
		t.Fatal(err)
	}
}

func Test_wca_Sync_with_db(t *testing.T) {
	s := syncer{
		DbPath:   "/Users/guojia/data/cubingPro/wca_db",
		SyncPath: "/Users/guojia/data/cubingPro/wca_db/sync_path",
		DbURL:    "root@tcp(127.0.0.1:33036)/",
	}

	err := s.sync()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_syncer_syncAddIndex(t *testing.T) {
	s := syncer{
		DbPath:   "/Users/guojia/data/cubingPro/wca_db",
		SyncPath: "/Users/guojia/data/cubingPro/wca_db/sync_path",
		DbURL:    "root@tcp(127.0.0.1:33036)/",
	}

	err := s.syncAddIndex("wca_20251226", syncWcaDbIndex)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_wca_Sync1(t *testing.T) {
	_ = NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		true)

}
