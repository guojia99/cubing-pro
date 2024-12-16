package plugin

import (
	"fmt"
	svc2 "github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"testing"
)

func TestRecordPlugin_Do(t *testing.T) {
	svc, err := svc2.NewAPISvc("/home/guojia/worker/code/cube/cubing-pro/etc/server_local.yaml")

	if err != nil {
		t.Fatal(err)
	}

	r := &RecordPlugin{Svc: svc}

	out, err := r.Do(types.InMessage{
		Message: "记录-666",
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out.Message)

}
