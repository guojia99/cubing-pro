package tools

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

type randomValue struct {
	value  []string
	num    int
	repeat bool
}

var randomKeys = map[string]randomValue{
	"edge": {
		value: []string{"CD", "EF", "GH", "IJ", "KL", "MN", "OP", "QR", "ST", "WX", "YZ"},
		num:   2, repeat: false,
	},
	"corner": {
		value:  []string{"ABC", "DEF", "GIH", "MNW", "OPQ", "RST", "XYZ"},
		num:    2,
		repeat: false,
	},
	"xedge": {
		value: []string{"CD", "EF", "GH", "IJ", "KL", "MN", "OP", "QR", "ST", "WX", "YZ"},
		num:   2, repeat: true,
	},
	"xcorner": {
		value:  []string{"ABC", "DEF", "GIH", "MNW", "OPQ", "RST", "XYZ"},
		num:    2,
		repeat: true,
	},
	"default": {
		value: []string{
			"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
			"N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"},
		num:    2,
		repeat: false,
	},
}

type TRandom struct {
}

func (t *TRandom) ID() []string {
	return []string{
		"random", "随机",
		"选择f", "select_f", "选择", "select",
	}
}

func (t *TRandom) Help() string {
	return `
1. 输出随机的字母组合
2. 随机 3bf {edge | corner | xEdge | xCorner} {*n}
3. 随机 {A B C D E} 随机自定义
4. 选择 {A B C} 随机选择
5. 选择f xxx{A B C} xxx {D E F} 在特殊语句中选择
`
}

func (t *TRandom) Do(message types.InMessage) (*types.OutMessage, error) {

	if utils.ContainsString(message.Message, "random", "随机") {
		return t.random(message)
	}
	return t.rSelect(message)
}

func (t *TRandom) rSelect(message types.InMessage) (*types.OutMessage, error) {

	fString := utils.ContainsString(message.Message, "选择f", "select_f")
	msg := types.RemoveID(message.Message, t.ID())

	getList := func(sl []string) []string {
		var newList []string
		for _, v := range sl {
			v = strings.TrimLeft(v, " ")
			if len(v) == 0 {
				continue
			}
			if v == " " {
				continue
			}
			newList = append(newList, v)
		}
		newList = utils.ShuffledCopy(newList, false)
		return newList
	}
	if !fString {
		sl := strings.Split(msg, " ")
		if len(sl) == 0 {
			return message.NewOutMessage("空空如也"), nil
		}
		return message.NewOutMessage(getList(sl)[0]), nil
	}
	fRe := regexp.MustCompile(`\{([^}]+)\}`)

	for {
		match := fRe.FindStringSubmatch(msg)
		if match == nil {
			break
		}
		options := strings.Split(match[1], " ")
		msg = strings.Replace(msg, match[0], getList(options)[0], 1)
	}
	return message.NewOutMessage(msg), nil
}

func (t *TRandom) random(message types.InMessage) (*types.OutMessage, error) {
	msg := types.RemoveID(message.Message, t.ID())
	lists := t.randomWithMsg(msg)
	if len(lists) == 1 {
		return message.NewOutMessage(lists...), nil
	}
	out := message.NewOutMessage()

	for i, l := range lists {
		out.AddMessagef("%d. %s\n", i+1, l)
	}
	return out, nil
}

func (t *TRandom) randomWithMsg(msg string) []string {
	msg = strings.ReplaceAll(msg, " ", "")
	var num = 1
	if strings.Contains(msg, "*") {
		numStrIdx := strings.Index(msg, "*")
		var err error
		if num, err = strconv.Atoi(msg[numStrIdx+1:]); err != nil {
			num = 1
		}
		msg = msg[:numStrIdx]
	}
	if num > 100 {
		num = 100
	}

	var val randomValue
	if len(msg) == 0 || len(strings.ReplaceAll(msg, " ", "")) == 0 || strings.Index(msg, "*") == 0 {
		val = randomKeys["default"]
	} else if strings.Contains(msg, "3bf") {
		for _, k := range []string{
			"xEdge", "xCorner", "edge", "corner",
		} {
			if strings.Contains(msg, k) {
				val = randomKeys[k]
				break
			}
		}
	} else {
		val = randomValue{
			value:  strings.Split(msg, ""),
			num:    2,
			repeat: false,
		}
	}

	var outs []string
	for i := 0; i < num; i++ {
		s := utils.ShuffledCopy(val.value, val.repeat)
		var l = ""
		for x, v := range s {
			if x%val.num == 0 {
				l += " "
			}
			if len(v) <= 1 {
				l += v
				continue
			}

			xx := utils.ShuffledCopy(strings.Split(v, ""), false)
			l += xx[0]
		}
		outs = append(outs, l)
	}
	return outs
}
