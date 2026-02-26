package wca

import (
	"testing"
)

func Test_wca_Sync(t *testing.T) {
	s := syncer{
		DbPath:   "/home/guojia/cubingPro/wca_db",
		SyncPath: "/home/guojia/cubingPro/wca_db/sync_path",
		DbURL:    "root@tcp(127.0.0.1:33306)/",
	}
	if err := s.init(); err != nil {
		t.Fatal(err)
	}
}

func Test_wca_Sync_with_db(t *testing.T) {
	s := syncer{
		DbPath:   "/home/guojia/cubingPro/wca_db",
		SyncPath: "/home/guojia/cubingPro/wca_db/sync_path",
		DbURL:    "root@tcp(127.0.0.1:33306)/",
	}

	err := s.sync()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_syncer_syncAddIndex(t *testing.T) {
	s := syncer{
		DbPath:   "/home/guojia/cubingPro/wca_db",
		SyncPath: "/home/guojia/cubingPro/wca_db/sync_path",
		DbURL:    "root@tcp(127.0.0.1:33306)/",
	}

	err := s.syncAddIndex("wca_20251226", syncWcaDbIndex)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_wca_Sync1(t *testing.T) {
	_ = NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path",
		true)

}
