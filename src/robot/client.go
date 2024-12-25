package robot

import (
	"context"

	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/robots"
	"github.com/guojia99/cubing-pro/src/robot/robots/plugin"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type Client struct {
	Svc *svc.Svc

	robots []types.Robot
}

func NewRobot(svc *svc.Svc) *Client {
	cli := &Client{
		Svc:    svc,
		robots: []types.Robot{},
	}

	for _, cq := range svc.Cfg.Robot.CQHttpBot {
		if cq.Enable {
			cli.robots = append(cli.robots, robots.NewCqHttps(&cq))
		}
	}
	for _, gq := range svc.Cfg.Robot.QQBot {
		if gq.Enable {
			cli.robots = append(cli.robots, robots.NewQQBot(&gq, context.Background()))
		}
	}

	return cli
}

func (c *Client) Run(ctx context.Context) error {
	plugins := plugin.NewPlugins(c.Svc)
	for _, bot := range c.robots {
		go robots.RunRobot(ctx, bot, plugins)
	}

	select {
	case <-ctx.Done():
		return nil
	}
}
