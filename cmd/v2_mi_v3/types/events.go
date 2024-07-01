package main

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

// todo 补充 icon
var wcaEvents = []event.Event{
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "333",
		},
		Name:          "333",
		OtherNames:    "三阶;三速;",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "222",
		},
		Name:          "222",
		OtherNames:    "二阶",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "444",
		},
		Name:          "444",
		OtherNames:    "四阶;",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "555",
		},
		Name:          "555",
		OtherNames:    "五阶",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	}, {
		StringIDModel: basemodel.StringIDModel{
			ID: "666",
		},
		Name:          "666",
		OtherNames:    "六阶",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType3roundsAvg,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "777",
		},
		Name:          "777",
		OtherNames:    "七阶;",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType3roundsAvg,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "pyram",
		},
		Name:          "pyram",
		OtherNames:    "塔;金字塔;py;",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "skewb",
		},
		Name:          "skewb",
		OtherNames:    "斜转;sk;",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "minx",
		},
		Name:          "megaminx",
		OtherNames:    "五魔;mx;明克斯",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "sq1",
		},
		Name:          "SQ-1",
		OtherNames:    "sq1;sq",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "clock",
		},
		Name:          "clock",
		OtherNames:    "魔表;ck",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "333bf",
		},
		Name:          "333bf",
		OtherNames:    "三盲;3bf",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType3roundsBest,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "444bf",
		},
		Name:          "444bf",
		OtherNames:    "4bf;四盲",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType3roundsBest,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "555bf",
		},
		Name:          "555bf",
		OtherNames:    "五盲;5bf",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "333mbf",
		},
		Name:          "333mbf",
		OtherNames:    "三阶多盲;3mbf;多盲",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteTypeRepeatedly,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "333oh",
		},
		Name:          "333oh",
		OtherNames:    "3oh;三单;",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	}, {
		StringIDModel: basemodel.StringIDModel{
			ID: "333fm",
		},
		Name:          "333fm",
		OtherNames:    "最少步;fmc;3fm",
		Class:         "wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         true,
		BaseRouteType: event.RouteType3roundsAvg,
	},

	// 已被取消的项目
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "333ft",
		},
		Name:          "333ft",
		OtherNames:    "脚拧;3ft",
		Class:         "old_wca",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         false,
		BaseRouteType: event.RouteType5RoundsAvgHT,
	},

	// 其他项目
}

var otherEvents = []event.Event{
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "333mbf_unlimited",
		},
		Name:          "333mbf Unlimited",
		OtherNames:    "3mbfu;无限多盲;不限时多盲;旧规则多盲",
		Class:         "盲拧",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         false,
		BaseRouteType: event.RouteTypeRepeatedly,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "2_7relay",
		},
		Name:          "2-7relay",
		OtherNames:    "27relay;正阶连拧;二至七连拧",
		Class:         "连拧",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         false,
		BaseRouteType: event.RouteType1rounds,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "alien_relay",
		},
		Name:          "alien relay",
		OtherNames:    "异形连拧", // 塔斜表五Q
		Class:         "连拧",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         false,
		BaseRouteType: event.RouteType1rounds,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "27alien_relay",
		},
		Name:          "27alien_relay",
		OtherNames:    "全项目连拧",
		Class:         "连拧",
		IsComp:        true,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         false,
		BaseRouteType: event.RouteType1rounds,
	},
}

var notCubes = []event.Event{
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "lunch",
		},
		Name:          "lunch",
		OtherNames:    "用餐",
		Class:         "other",
		IsComp:        false,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         false,
		BaseRouteType: 0,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "registration",
		},
		Name:          "registration",
		OtherNames:    "签到;注册",
		Class:         "other",
		IsComp:        false,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         false,
		BaseRouteType: 0,
	},
	{
		StringIDModel: basemodel.StringIDModel{
			ID: "submission",
		},
		Name:          "Puzzle Submission",
		OtherNames:    "提交魔方",
		Class:         "other",
		IsComp:        false,
		Icon:          "",
		IconBase64:    "",
		IsWCA:         false,
		BaseRouteType: 0,
	},
}
