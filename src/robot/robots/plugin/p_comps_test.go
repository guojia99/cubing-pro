package plugin

import (
	"fmt"
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

func TestCompsPlugin__getComps(t *testing.T) {
	groupID := "966F739CD99E3B4B3437BCB738400CB9"
	sv, err := svc.NewAPISvc(
		"/Users/guojia/worker/code/cube/cubing-pro/local/server_local_dev.yaml",
		false,
		false,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	c := &CompsPlugin{
		Svc: sv,
	}
	got, got1, _, err := c._getComps(
		types.InMessage{
			Message: "比赛",
			GroupID: groupID,
		},
		1,
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(got)
	fmt.Println(got1)
}

func TestCompsPlugin_Do(t *testing.T) {
	groupID := "966F739CD99E3B4B3437BCB738400CB9"
	sv, err := svc.NewAPISvc(
		"/Users/guojia/worker/code/cube/cubing-pro/local/server_local_dev.yaml",
		false,
		false,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	c := &CompsPlugin{
		Svc: sv,
	}

	messages := []string{
		"打乱 333",
		"打乱 333 2",
		"打乱 333 复赛第1轮",
		"赛果-300 333 2",
	}

	for _, msg := range messages {
		t.Run(msg, func(t *testing.T) {
			out, err := c.Do(types.InMessage{
				Message: msg,
				GroupID: groupID,
			})
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(msg)
			fmt.Println(out)
		})
	}
}
