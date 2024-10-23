package plugin

import (
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

func NewPlugins(svc *svc.Svc) []types.Plugin {
	return []types.Plugin{
		&TryPlugin{Svc: svc},
		&CompsPlugin{Svc: svc},
		&PlayerPlugin{Svc: svc},
	}
}
