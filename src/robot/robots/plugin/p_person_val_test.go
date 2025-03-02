package plugin

import (
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

func TestPersonValPlugin_init(t *testing.T) {
	svc2 := &svc.Svc{
		Cfg: configs.Config{
			Robot: configs.RobotConfig{
				PersonValPath: "/home/guojia/worker/code/cube/cubing-pro/static/personValPath",
			},
		},
	}

	p := &PersonValPlugin{Svc: svc2}
	p.init()
}
