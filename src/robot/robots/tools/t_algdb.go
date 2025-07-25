package tools

import (
	"fmt"
	"strings"
	"sync"

	"github.com/guojia99/cubing-pro/src/internel/algdb"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type TAlgDB struct {
	Svc    *svc.Svc
	dbs    []algdb.AlgDB
	dbsMap map[string]algdb.AlgDB

	one sync.Once
}

func (t *TAlgDB) init() {
	t.dbs = []algdb.AlgDB{
		algdb.NewSQ1CspDB(t.Svc.Cfg.GlobalConfig.AlgPath),
		algdb.NewBldDB(t.Svc.Cfg.GlobalConfig.AlgPath),
		algdb.NewCube222(t.Svc.Cfg.GlobalConfig.AlgPath),
		algdb.NewCubePy(t.Svc.Cfg.GlobalConfig.AlgPath),
		algdb.NewCube333(t.Svc.Cfg.GlobalConfig.AlgPath),
	}

	t.dbsMap = make(map[string]algdb.AlgDB)
	for _, db := range t.dbs {
		for _, id := range db.ID() {
			t.dbsMap[id] = db
		}
	}
}

func (t *TAlgDB) ID() []string {
	t.one.Do(t.init)
	return []string{"alg", "公式"}
}

func (t *TAlgDB) Help() string {
	out := "1. 输入： 公式 xxxx 可查询某些公式\n"
	idx := 2
	for i, db := range t.dbs {
		out += fmt.Sprintf("%d. %s\n", idx+i, db.Help())
	}
	return out
}

func (t *TAlgDB) Do(message types.InMessage) (*types.OutMessage, error) {
	msg := types.RemoveID(message.Message, t.ID())

	sp := strings.Split(strings.TrimLeft(msg, " "), " ")
	if len(sp) == 0 {
		return message.NewOutMessage(t.Help()), nil
	}
	key := sp[0]
	db, ok := t.dbsMap[key]
	if !ok {
		return message.NewOutMessage(t.Help()), nil
	}
	config := db.BaseConfig()

	output, img, err := db.Select(msg, config)
	if err != nil {
		return message.NewOutMessagef("%+v", err), nil
	}

	if img != "" {
		return message.NewOutMessageWithImage(output, img), nil
	}
	return message.NewOutMessage(output), nil

}
