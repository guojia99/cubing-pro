package tools

import (
	"fmt"
	"log"
	"strings"

	resultDB "github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/models"
	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/utils"
	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/internel/wca"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type TWca struct {
	DB    *gorm.DB
	Cache *cache.Cache
}

func (t *TWca) ID() []string {
	wcaL := []string{"Wca", "wca", "WCA"}
	pk := []string{"pk", "PK", "-PK", "-pk",
		"pkAll", "-pkAll", "PKAll", "-PKAll"}
	cx := []string{"超炫", "out", "cx", "CX"}

	var out []string
	for _, w := range wcaL {
		out = append(out, w)
		for _, p := range pk {
			out = append(out, fmt.Sprintf("%s%s", w, p))
		}
		for _, c := range cx {
			out = append(out, fmt.Sprintf("%s%s", w, c))
		}
	}

	return out
}

func (t *TWca) Help() string {
	out := `1. 输入 WCA {WcaID} 可查询选手成绩
2. 输入 WCA-PK {WCAID-1}-{WCAID-2} 可对比成绩（只有双方都有的项目), WCA-PKAll可展示全部项目
3. 输入 WCA-超炫 {WCAID-1}-{WCAID-2} 可对比成绩后， 列出1超炫2需要进步多少
`
	return out
}

var wcaEventsList = []string{
	"333",
	"222",
	"444",
	"555",
	"666",
	"777",
	"333bf",
	"333fm",
	"333oh",
	"clock",
	"minx",
	"pyram",
	"skewb",
	"sq1",
	"444bf",
	"555bf",
	"333mbf",
}

var wcaEventsCnMap = map[string]string{
	"333":    "三阶",
	"222":    "二阶",
	"444":    "四阶",
	"555":    "五阶",
	"666":    "六阶",
	"777":    "七阶",
	"333bf":  "三盲",
	"333fm":  "最少步",
	"333oh":  "单手",
	"clock":  "魔表",
	"minx":   "五魔",
	"pyram":  "金字塔",
	"skewb":  "斜转",
	"sq1":    "SQ-1",
	"444bf":  "四盲",
	"555bf":  "五盲",
	"333mbf": "多盲",
}

func (t *TWca) Do(message types.InMessage) (*types.OutMessage, error) {
	msg := strings.ToLower(message.Message)
	slices := utils2.Split(msg, " ")
	if len(slices) <= 1 {
		return message.NewOutMessage(t.Help()), nil
	}

	key := slices[0]

	switch key {
	case "wca-pkall", "wcapkall":
		return t.handlerPkDoublePersonResult(message, true)
	case "wca-pk", "wcapk":
		return t.handlerPkDoublePersonResult(message, false)
	case "wca":
		return t.handlerGetPersonResult(message)
	case "wca超炫", "wcacx":
		return t.handlerCxDoublePersonResult(message)
	default:
		return message.NewOutMessage(t.Help()), nil
	}
}

func (t *TWca) getPersonResult(msg string) (*models.PersonBestResults, string, error) {
	pes, err := wca.ApiSearchPersons(msg)
	if err != nil {
		log.Printf("wca api search persons err: %v", err)
		return nil, "", fmt.Errorf("获取%s选手失败", msg)
	}
	if pes.Total == 0 || len(pes.Rows) == 0 {
		return nil, "", fmt.Errorf("查询不到选手%s", msg)
	}
	if pes.Total >= 2 {
		out := fmt.Sprintf("查询到%d位选手, 请输入WcaID查询: \n", pes.Total)
		for _, row := range pes.Rows {
			out += fmt.Sprintf("%s %s\n", row.WcaId, row.Name)
		}
		return nil, out, fmt.Errorf("查询到多位选手符合%s, 请使用WCAID查询", msg)
	}
	personWCAID := pes.Rows[0].WcaId

	result, err := wca.GetWcaResultWithDbAndAPI(t.DB, personWCAID)
	if err != nil {
		return nil, "", fmt.Errorf("选手%s成绩查询错误", personWCAID)
	}
	return result, "", nil
}

func (t *TWca) handlerGetPersonResult(message types.InMessage) (*types.OutMessage, error) {
	msg := types.RemoveID(message.Message, t.ID())
	msg = utils2.ReplaceAll(msg, "", " ")

	result, mutilMsg, err := t.getPersonResult(msg)
	if err != nil {
		if len(mutilMsg) > 0 {
			return message.NewOutMessage(mutilMsg), nil
		}
		return message.NewOutMessage(err.Error()), nil
	}

	out := result.PersonName + "\n"
	out += result.WCAID + "\n"
	for _, ev := range wcaEventsList {
		b, hasB := result.Best[ev]
		if !hasB {
			continue
		}
		a, hasA := result.Avg[ev]
		if hasA {
			out += fmt.Sprintf("%s %s | %s\n", wcaEventsCnMap[ev], b.BestStr, a.AverageStr)
		} else {
			out += fmt.Sprintf("%s %s\n", wcaEventsCnMap[ev], b.BestStr)
		}
	}
	return message.NewOutMessage(out), nil
}

const (
	startEmpty = " "
	starWin    = "🌟"
	starP1     = "☆"
	starP2     = "★"
)

func (t *TWca) pk(p1 *models.Results, p2 *models.Results, best bool, full bool) (p1Count, p2Count int, msg string) {
	if p1 == nil && p2 == nil {
		return 0, 0, ""
	}

	// 两边都有成绩才对比
	if !full && (p1 == nil || p2 == nil) {
		return 0, 0, ""
	}

	if p1 != nil && p2 == nil {
		p1ResultStr := p1.BestStr
		if !best {
			p1ResultStr = p1.AverageStr
		}
		return 1, 0, fmt.Sprintf("%s %s|| -", starP1, p1ResultStr)
	}

	if p1 == nil {
		p2ResultStr := p2.BestStr
		if !best {
			p2ResultStr = p2.AverageStr
		}
		return 0, 1, fmt.Sprintf("%s - || %s %s", startEmpty, p2ResultStr, starP2)
	}

	p1ResultStr, p1Result, p2ResultStr, p2Result := p1.BestStr, p1.Best, p2.BestStr, p2.Best
	if !best {
		p1ResultStr, p1Result, p2ResultStr, p2Result = p1.AverageStr, p1.Average, p2.AverageStr, p2.Average
	}

	if p1.EventId == "333mbf" {
		p1Solved, p1Attempted, p1Senconds, _ := utils.Get333MBFResult(p1.Best)
		p2Solved, p2Attempted, p2Senconds, _ := utils.Get333MBFResult(p2.Best)

		p1Num := p1Solved - (p1Attempted - p1Solved)
		p2Num := p2Solved - (p2Attempted - p2Solved)

		if p1Num == p2Num { // 分数相同时
			if p1Senconds == p2Senconds { // 时间相同
				p1Result, p2Result = 0, 0
			} else if p1Senconds < p2Senconds {
				p1Result, p2Result = -1, 1 // p1 更快
			} else {
				p1Result, p2Result = 1, -1 // p2更快
			}
		} else if p1Num > p2Num {
			p1Result, p2Result = -1, 1 // p1更多
		} else {
			p1Result, p2Result = 1, -1 // p2更多
		}
	}

	// 对比
	if p1Result == p2Result {
		return 1, 1, fmt.Sprintf("%s %s || %s %s", starP1, p1ResultStr, p2ResultStr, starP2)
	}
	if p1Result < p2Result {
		return 1, 0, fmt.Sprintf("%s %s || %s", starP1, p1ResultStr, p2ResultStr)
	}
	return 0, 1, fmt.Sprintf("%s %s || %s %s", startEmpty, p1ResultStr, p2ResultStr, starP2)
}

func (t *TWca) getDoublePerson(message types.InMessage) (*models.PersonBestResults, *models.PersonBestResults, error) {
	msg := types.RemoveID(message.Message, t.ID())
	slices := utils2.Split(msg, "-")
	var person1, person2 string
	if len(slices) == 2 {
		person1 = slices[0]
		person2 = slices[1]
	} else if strings.Contains(msg, "/") || strings.Contains(msg, "\\") || strings.Contains(msg, "VS") {
		msg = utils2.ReplaceAll(msg, "VS", "/", "\\")
		slices = utils2.Split(msg, "VS")
		person1 = slices[0]
		person2 = slices[1]
	} else {
		return nil, nil, fmt.Errorf("%+v", t.Help())
	}

	person1Result, _, err := t.getPersonResult(person1)
	if err != nil {
		return nil, nil, err
	}
	person2Result, _, err := t.getPersonResult(person2)
	if err != nil {
		return nil, nil, err
	}
	return person1Result, person2Result, nil
}

func (t *TWca) handlerPkDoublePersonResult(message types.InMessage, full bool) (*types.OutMessage, error) {

	// 对比两个人的
	person1Count := 0
	person2Count := 0

	person1Result, person2Result, err := t.getDoublePerson(message)
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}

	var out string
	out += fmt.Sprintf("%s PK %s\n", person1Result.PersonName, person2Result.PersonName)

	for _, ev := range wcaEventsList {
		var p1BestResult *models.Results
		var p2BestResult *models.Results
		if v, ok := person1Result.Best[ev]; ok {
			p1BestResult = &v
		}
		if v, ok := person2Result.Best[ev]; ok {
			p2BestResult = &v
		}

		if p1BestResult == nil && p2BestResult == nil {
			continue
		}

		p1, p2, pkMsg := t.pk(p1BestResult, p2BestResult, true, full)
		if pkMsg == "" {
			continue
		}

		out += fmt.Sprintf("%s %s\n", wcaEventsCnMap[ev], pkMsg)
		person1Count += p1
		person2Count += p2

		var p1AvgResult *models.Results
		var p2AvgResult *models.Results
		if v, ok := person1Result.Avg[ev]; ok {
			p1AvgResult = &v
		}
		if v, ok := person2Result.Avg[ev]; ok {
			p2AvgResult = &v
		}
		if p1AvgResult == nil && p2AvgResult == nil {
			continue
		}
		p1, p2, pkMsg = t.pk(p1AvgResult, p2AvgResult, false, full)
		if pkMsg == "" {
			continue
		}
		out += fmt.Sprintf("%s %s\n", strings.Repeat(" ", len(wcaEventsCnMap[ev])/2), pkMsg)
		person1Count += p1
		person2Count += p2
	}
	out += "\n"
	if person1Count == person2Count {
		out += fmt.Sprintf("结果: 平手 %d:%d", person1Count, person2Count)
	} else if person1Count > person2Count {
		out += fmt.Sprintf("结果: (%s)%d:%d\n", starWin, person1Count, person2Count)
		out += fmt.Sprintf("%s胜利 \n", person1Result.PersonName)

	} else {
		out += fmt.Sprintf("结果:%d:%d(%s)\n", person1Count, person2Count, starWin)
		out += fmt.Sprintf("%s胜利 \n", person2Result.PersonName)

	}

	return message.NewOutMessage(out), nil
}

func (t *TWca) cx(p1 *models.Results, p2 *models.Results, best bool) (p1Wined, p1WillWin, notWin string) {
	if p1 == nil && p2 == nil {
		return
	}

	evs := "平均"
	start := starP1
	if best {
		evs = "单次"
		start = starP2
	}

	if p1 != nil && p2 == nil {
		p1ResultStr := p1.BestStr
		if !best {
			p1ResultStr = p1.AverageStr
		}
		return fmt.Sprintf("%s %s项目%s已经完全超炫他了! 你: %s", start, wcaEventsCnMap[p1.EventId], evs, p1ResultStr), "", ""
	}

	if p1 == nil {
		p2ResultStr := p2.BestStr
		if !best {
			p2ResultStr = p2.AverageStr
		}
		return "", fmt.Sprintf("%s %s项目%s被他完全超炫! 他: %s", start, wcaEventsCnMap[p2.EventId], evs, p2ResultStr), ""
	}

	p1ResultStr, p1Result, p2ResultStr, p2Result := p1.BestStr, p1.Best, p2.BestStr, p2.Best
	if !best {
		p1ResultStr, p1Result, p2ResultStr, p2Result = p1.AverageStr, p1.Average, p2.AverageStr, p2.Average
	}
	wcaP1Result, wcaP2Result := utils.WCAResultIntToSeconds(p1Result, p1.EventId), utils.WCAResultIntToSeconds(p2Result, p2.EventId)

	if p1.EventId == "333mbf" {
		p1Solved, p1Attempted, p1Senconds, _ := utils.Get333MBFResult(p1.Best)
		p2Solved, p2Attempted, p2Senconds, _ := utils.Get333MBFResult(p2.Best)

		p1Num := p1Solved - (p1Attempted - p1Solved)
		p2Num := p2Solved - (p2Attempted - p2Solved)

		if p1Num == p2Num { // 分数相同时
			if p1Senconds == p2Senconds { // 时间相同
				p1Result, p2Result = 0, 0
			} else if p1Senconds < p2Senconds {
				p1Result, p2Result = -1, 1 // p1 更快
			} else {
				p1Result, p2Result = 1, -1 // p2更快
			}
		} else if p1Num > p2Num {
			p1Result, p2Result = -1, 1 // p1更多
		} else {
			p1Result, p2Result = 1, -1 // p2更多
		}
	}

	if p1Result == p2Result {
		return "", fmt.Sprintf("%s %s%s你们打平手了, 你们:%s, 你只需要进步0.01秒", start, wcaEventsCnMap[p1.EventId], evs, p1ResultStr), ""
	}

	if p1Result < p2Result {
		if p1.EventId == "333mbf" {
			return fmt.Sprintf("%s %s 你已经超炫了他, 你:%s, 他:%s", start, wcaEventsCnMap[p1.EventId], p1ResultStr, p2ResultStr), "", ""
		}
		return fmt.Sprintf("%s %s%s, 你:%s, 他:%s, 你超炫了他%s", start, wcaEventsCnMap[p1.EventId], evs, p1ResultStr, p2ResultStr, resultDB.TimeParserF2S(wcaP2Result-wcaP1Result)), "", ""
	}

	if p2.EventId == "333mbf" {
		return "", fmt.Sprintf("%s %s 你被他超炫了,你:%s, 他:%s", start, wcaEventsCnMap[p2.EventId], p1ResultStr, p2ResultStr), ""
	}
	return "", fmt.Sprintf("%s %s%s你被他超炫, 你:%s, 他:%s, 你需要进步:%s", start, wcaEventsCnMap[p2.EventId], evs, p1ResultStr, p2ResultStr, resultDB.TimeParserF2S(wcaP1Result-wcaP2Result)), ""
}

func (t *TWca) handlerCxDoublePersonResult(message types.InMessage) (*types.OutMessage, error) {
	person1Result, person2Result, err := t.getDoublePerson(message)
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}

	var p1WinP2Results []string
	var p1WillWinP2Results []string
	var notWineds []string

	for _, ev := range wcaEventsList {
		// 单次
		var p1BestResult *models.Results
		var p2BestResult *models.Results
		if v, ok := person1Result.Best[ev]; ok {
			p1BestResult = &v
		}
		if v, ok := person2Result.Best[ev]; ok {
			p2BestResult = &v
		}

		if p1BestResult == nil && p2BestResult == nil {
			continue
		}
		p1W, p2W, notWin := t.cx(p1BestResult, p2BestResult, true)
		if p1W != "" {
			p1WinP2Results = append(p1WinP2Results, p1W)
		}
		if p2W != "" {
			p1WillWinP2Results = append(p1WillWinP2Results, p2W)
		}
		if notWin != "" {
			notWineds = append(notWineds, notWin)
		}

		// 平均
		var p1AvgResult *models.Results
		var p2AvgResult *models.Results
		if v, ok := person1Result.Avg[ev]; ok {
			p1AvgResult = &v
		}
		if v, ok := person2Result.Avg[ev]; ok {
			p2AvgResult = &v
		}
		if p1AvgResult == nil && p2AvgResult == nil {
			continue
		}
		p1W, p2W, notWin = t.cx(p1AvgResult, p2AvgResult, false)
		if p1W != "" {
			p1WinP2Results = append(p1WinP2Results, p1W)
		}
		if p2W != "" {
			p1WillWinP2Results = append(p1WillWinP2Results, p2W)
		}
		if notWin != "" {
			notWineds = append(notWineds, notWin)
		}
	}

	out := "\n"

	if len(p1WinP2Results) > 0 {
		out += "\n =============================\n"
		out += fmt.Sprintf("%s 超炫 %s的项目:\n", person1Result.PersonName, person2Result.PersonName)
		for _, p1WinP2Result := range p1WinP2Results {
			out += p1WinP2Result + "\n"
		}
	}

	if len(p1WillWinP2Results) > 0 {
		out += "\n =============================\n"
		out += fmt.Sprintf("%s 被%s超炫的项目: \n", person1Result.PersonName, person2Result.PersonName)
		for _, p1WillWinP2Result := range p1WillWinP2Results {
			out += p1WillWinP2Result + "\n"
		}
	}

	if len(notWineds) > 0 {
		out += "\n =============================\n"
		out += fmt.Sprintf("%s和%s打平手的项目: \n", person1Result.PersonName, person2Result.PersonName)
		for _, notWined := range notWineds {
			out += notWined + "\n"
		}
	}

	return message.NewOutMessage(out), nil
}
