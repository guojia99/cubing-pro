package algs

var CubeKeyList = []string{
	"222", "special", "333", "333oh", "sq1", "big cube",
}
var algsDataKey = map[string][]string{
	"222": {
		"2x2-OLL-Trainer",
		"2x2-PBL-Trainer",
		"2x2-EG-Trainer",
		"2x2-TCLL-Trainer",
		"2x2-TEG-Trainer",
		"2x2-FH-Trainer",
		"2x2-LS-Trainer",
	},
	"special": {
		"Pyraminx-L4E-Trainer",
		"Skewb-NS2-Trainer",
		"Megaminx-OLL-Trainer",
		"Megaminx-PLL-Trainer",
		"Octaminx-L3T-Trainer",
		"Octaminx-TCP-Trainer",
	},
	"333": {
		"3x3-F2L-Trainer",
		"3x3-OLL-Trainer",
		"3x3-PLL-Trainer",
		"3x3-COLL-Trainer",
		"3x3-CMLL-Trainer",
		"3x3-ZBLL-Trainer",
		"3x3-ZBLS-Trainer",
	},
	"333oh": {
		"3x3-OH-CMLL-Trainer",
		"3x3-OH-OLL-Trainer",
		"3x3-OH-PLL-Trainer",
		"3x3-OH-ZBLL-Trainer",
	},

	"sq1": {
		"SQ1-CS-Trainer",
		"SQ1-CO-Trainer",
		"SQ1-EO-Trainer",
		"Sq1-CPEP-Trainer",
		"Sq1-OBL-Trainer",
		"Sq1-PBL-Trainer",
	},
	"big cube": {
		"4x4-PLLP-Trainer",
		"5x5-L2E-Trainer",
	},
}

var algsNameMap = map[string]string{
	"2x2-OLL-Trainer":  "COLL",
	"2x2-PBL-Trainer":  "PBL",
	"2x2-EG-Trainer":   "EG",
	"2x2-FH-Trainer":   "FH",
	"2x2-LS-Trainer":   "LS",
	"2x2-TCLL-Trainer": "TCLL",
	"2x2-TEG-Trainer":  "TEG",

	"3x3-OLL-Trainer":  "OLL",
	"3x3-PLL-Trainer":  "PLL",
	"3x3-COLL-Trainer": "COLL",
	"3x3-CMLL-Trainer": "CMLL",
	"3x3-ZBLL-Trainer": "ZBLL",
	"3x3-ZBLS-Trainer": "ZBLS",
	"3x3-F2L-Trainer":  "F2L",

	"3x3-OH-CMLL-Trainer": "CMLL",
	"3x3-OH-OLL-Trainer":  "OLL",
	"3x3-OH-PLL-Trainer":  "PLL",
	"3x3-OH-ZBLL-Trainer": "ZBLL",

	"SQ1-CS-Trainer":   "CS",
	"SQ1-CO-Trainer":   "CO",
	"SQ1-EO-Trainer":   "EO",
	"Sq1-CPEP-Trainer": "CPEP",
	"Sq1-OBL-Trainer":  "OBL",
	"Sq1-PBL-Trainer":  "PBL",

	"Megaminx-OLL-Trainer": "Megaminx OLL",
	"Megaminx-PLL-Trainer": "Megaminx PLL",
	"Skewb-NS2-Trainer":    "Skewb NS",
	"Pyraminx-L4E-Trainer": "Pyraminx L4E",
	"Octaminx-L3T-Trainer": "FTO L3T",
	"Octaminx-TCP-Trainer": "FTO TCP",

	"5x5-L2E-Trainer":  "L2E",
	"4x4-PLLP-Trainer": "PLLP",
}
