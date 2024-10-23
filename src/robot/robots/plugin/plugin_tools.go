package plugin

import (
	"fmt"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"strings"
	"time"
)

func PluginsMap(plugins []types.Plugin) map[string]types.Plugin {
	var out = make(map[string]types.Plugin)
	for _, plugin := range plugins {
		for _, id := range plugin.ID() {
			if _, ok := out[id]; ok {
				fmt.Printf("载入Plugin %s已经存在\n", id)
				continue
			}
			out[id] = plugin
		}
	}
	return out
}

func RemoveID(message string, id []string) string {
	for _, i := range id {
		message = strings.TrimLeft(message, i)
	}
	return message
}

const MaxKeyLength = 16

func findSubSeq(input string) []string {
	var cache [][]rune
	msg := []rune(input)

	for i := 0; i < len(msg)+1; i++ {
		cache = append(cache, msg[:i])
		if i > MaxKeyLength {
			break
		}
	}

	var out []string
	for _, val := range cache {
		out = append(out, string(val))
	}
	return out
}

func CheckPrefix(msg string, pluginMap map[string]types.Plugin) (string, types.Plugin) {
	for _, key := range findSubSeq(msg) {
		if p, ok := pluginMap[key]; ok {
			return key, p
		}
	}
	return "", nil
}

func GetEvents(svc *svc.Svc, EventMin string) []event.Event {
	key := "robot_events_key_" + EventMin
	if data, ok := svc.Cache.Get(key); ok {
		return data.([]event.Event)
	}

	var events []event.Event
	var evs = strings.Split(EventMin, ";")
	if len(EventMin) > 0 {
		svc.DB.Where("id in ?", evs).Order("idx").Find(&events)
	} else {
		svc.DB.Order("idx").Find(&events)
	}

	svc.Cache.Set(key, events, time.Second*60)
	return events
}
