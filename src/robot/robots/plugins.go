package robots

import (
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/robots/plugin"
	"github.com/guojia99/cubing-pro/src/robot/robots/tools"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"github.com/patrickmn/go-cache"
)

func NewPlugins(svc *svc.Svc) []types.Plugin {
	return []types.Plugin{
		&plugin.TryPlugin{Svc: svc},
		&plugin.CompsPlugin{Svc: svc},
		&plugin.PlayerPlugin{Svc: svc},
		&plugin.RecordPlugin{Svc: svc},
		&plugin.RankPlugin{Svc: svc},
		&plugin.PreResultPlugin{Svc: svc},
		&plugin.BindPlugin{Svc: svc},
		//&PersonValPlugin{Svc: svc},

		&tools.TRandom{},
		&tools.TAlgDB{Svc: svc},
		&tools.TScramble{Svc: svc},

		&tools.TWca{Cache: cache.New(cache.DefaultExpiration, cache.NoExpiration), DB: svc.DB},
	}
}
