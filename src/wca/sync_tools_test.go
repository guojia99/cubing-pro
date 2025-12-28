package wca

import (
	"fmt"
	"testing"
)

func Test_checkRemoteFileDate(t *testing.T) {

	ts, url, err := checkRemoteFileDate()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ts)
	fmt.Println(url)
}

func Test_downloadIfNeeded(t *testing.T) {

	p := "/home/guojia/worker/code/cube/cubing-pro"
	url := "https://assets.worldcubeassociation.org/export/results/WCA_export_v2_357_20251223T000007Z.sql.zip"

	tp, err := downloadIfNeeded(p, url)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tp)

}

func Test_extractZipToDb(t *testing.T) {
	filePath := "/home/guojia/worker/code/cube/cubing-pro/20251223.zip"
	dbPath := "/home/guojia/worker/code/cube/cubing-pro"
	targetDir, err := extractZipToDb(filePath, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(targetDir)
}
