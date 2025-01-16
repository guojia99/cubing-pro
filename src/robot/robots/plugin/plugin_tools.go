package plugin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
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
		out = append(out, utils.ReplaceAll(string(val), "", " ", " "))
	}
	return out
}

func CheckPrefix(msg string, pluginMap map[string]types.Plugin) (string, types.Plugin) {
	msg = strings.TrimLeft(msg, " ")
	msg = utils.ReplaceAll(msg, " ", "")

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
	var evs = utils.Split(EventMin, ";")
	if len(EventMin) > 0 {
		svc.DB.Where("id in ?", evs).Order("idx").Find(&events)
	} else {
		svc.DB.Order("idx").Find(&events)
	}

	svc.Cache.Set(key, events, time.Second*60)
	return events
}

func GetMessageEvent(evs []event.Event, msg string) (event.Event, string, int, error) {
	msg = utils.ReplaceAll(msg, "", "-")
	if len(msg) == 0 {
		return event.Event{}, "", 0, fmt.Errorf("找不到该项目")
	}

	split := utils.Split(msg, " ")
	var ev event.Event
	var round string
	if len(split) >= 1 {
		eStr := split[0]
		if idx := strings.IndexAny(eStr, "[("); idx >= 0 {
			round = eStr[idx:]
			round = utils.ReplaceAll(round, "", "(", "[", "]", ")")

			eStr = eStr[:idx]
		}

		for _, e := range evs {
			if len(ev.ID) > 0 {
				break
			}
			if eStr == e.Cn || eStr == e.Name || eStr == e.ID {
				ev = e
				break
			}
			for _, s := range utils.Split(e.OtherNames, ";") {
				if s == eStr {
					ev = e
					break
				}
			}
		}
	}

	var num = 1
	if len(split) >= 2 {
		nStr := utils.ReplaceAll(split[1], "", "[", "]", "(", ")")
		n, err := strconv.Atoi(nStr)
		if err == nil {
			num = n
		}
	}

	if ev.ID == "" {
		return ev, round, num, fmt.Errorf("找不到项目")
	}

	if num > 50 {
		num = 50
	}

	return ev, round, num, nil
}

func getUser(svc *svc.Svc, message types.InMessage) (user.User, error) {
	var usr user.User
	var err error
	if message.QQ != 0 {
		err = svc.DB.Where("qq = ?", message.QQ).First(&usr).Error
	} else if message.QQBot != "" {
		err = svc.DB.Where("qq_uni_id = ?", message.QQBot).First(&usr).Error
	}
	return usr, err
}
