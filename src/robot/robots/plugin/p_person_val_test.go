package plugin

import (
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func TestPersonValPlugin_init(t *testing.T) {
	svc2 := &svc.Svc{
		Cfg: svc.Config{
			Robot: svc.RobotConfig{
				PersonValPath: "/home/guojia/worker/code/cube/cubing-pro/static/personValPath",
			},
		},
	}

	p := &PersonValPlugin{Svc: svc2}
	p.init()
}
