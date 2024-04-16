package scramble

import "errors"

type TNoodleEvents struct {
	Event      string
	TNoodleKey string
	Keys       []string
}

var events = []TNoodleEvents{
	{
		Event:      "222",
		TNoodleKey: "222",
		Keys:       []string{"2", "222", "二阶", "2阶", "二阶魔方", "2x2x2"},
	},
	{
		Event:      "333",
		TNoodleKey: "333",
		Keys:       []string{"3", "333", "三阶", "3阶", "三阶魔方", "3x3x3"},
	},
	{
		Event:      "444",
		TNoodleKey: "444",
		Keys:       []string{"4", "444", "四阶", "4阶", "四阶魔方", "4x4x4"},
	},
	{
		Event:      "555",
		TNoodleKey: "555",
		Keys:       []string{"5", "555", "五阶", "5阶", "五阶魔方", "5x5x5"},
	},
	{
		Event:      "666",
		TNoodleKey: "666",
		Keys:       []string{"6", "666", "六阶", "6阶", "六阶魔方", "6x6x6"},
	},
	{
		Event:      "777",
		TNoodleKey: "777",
		Keys:       []string{"7", "777", "七阶", "7阶", "七阶魔方", "7x7x7"},
	},
	{
		Event:      "clock",
		TNoodleKey: "clock",
		Keys:       []string{"cl", "clock", "魔表", "表"},
	},
	{
		Event:      "minx",
		TNoodleKey: "minx",
		Keys:       []string{"mx", "minx", "五魔", "五魔方", "megaminx"},
	},
	{
		Event:      "pyram",
		TNoodleKey: "pyram",
		Keys:       []string{"py", "pyram", "金字塔", "塔"},
	},
	{
		Event:      "skewb",
		TNoodleKey: "skewb",
		Keys:       []string{"sk", "skewb", "斜转", "斜"},
	},
	{
		Event:      "sq1",
		TNoodleKey: "sq1",
		Keys:       []string{"sq", "sq1", "sq-1"},
	},
	{
		Event:      "333oh",
		TNoodleKey: "333",
		Keys:       []string{"333oh", "三单", "单手", "单"},
	},
	{
		Event:      "333fm",
		TNoodleKey: "333fm",
		Keys:       []string{"333fm", "最少步", "蜂蜜茶"},
	},
	{
		Event:      "333bf",
		TNoodleKey: "333",
		Keys:       []string{"333bf", "3bf", "三盲", "三阶盲拧", "盲拧"},
	},
	{
		Event:      "444bf",
		TNoodleKey: "444",
		Keys:       []string{"444bf", "4bf", "四盲", "四阶盲拧"},
	},
	{
		Event:      "555bf",
		TNoodleKey: "555",
		Keys:       []string{"555bf", "5bf", "五盲", "五阶盲拧"},
	},
	{
		Event:      "333mbf",
		TNoodleKey: "333",
		Keys:       []string{"333mbf", "多盲", "3mbf"},
	},
}

var eventsMap = func() map[string]TNoodleEvents {
	var out = make(map[string]TNoodleEvents)
	for _, event := range events {
		for _, key := range event.Keys {
			out[key] = event
		}
	}
	return out
}()

func TNoodleKey(in string) (string, error) {
	val, ok := eventsMap[in]
	if ok {
		return val.TNoodleKey, nil
	}
	return "", errors.New("unknown key")
}
