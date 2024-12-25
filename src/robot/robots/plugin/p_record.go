package plugin

import (
	"fmt"
	"sort"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"github.com/guojia99/go-tables/table"
)

type RecordPlugin struct {
	Svc *svc.Svc
}

var _ types.Plugin = &RecordPlugin{}

func (r *RecordPlugin) ID() []string {
	return []string{"record", "记录", "g_record", "群记录"}
}

func (r *RecordPlugin) Help() string {
	return "获取记录列表"
}

type tableRecord struct {
	//Time   string `table:"时间"`
	Idx      int    `table:"-"`
	ResultID int    `table:"-"`
	Rank     int    `table:"序号"`
	Event    string `table:"项目"`
	Name     string `table:"选手"`
	Time     string `table:"-"`
	Single   string `table:"单次"`
	Avg      string `table:"平均"`
}

func (r *RecordPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	evs := GetEvents(r.Svc, "")
	msg := RemoveID(message.Message, r.ID())
	ev, _, num, err := GetMessageEvent(evs, msg)
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}

	var rc_t = result.RecordTypeWithCubingPro
	//if strings.Contains(message.Message, "群记录") || strings.Contains(message. Message, "g_record"){
	//	rc_t = result.RecordTypeWithGroup
	//}
	var records []result.Record
	r.Svc.DB.Where("d_type = ?", rc_t).Where("event_id = ?", ev.ID).Find(&records)

	if len(records) == 0 {
		return message.NewOutMessage("该项目无记录"), nil
	}

	var out = fmt.Sprintf("%s 记录列表\n", ev.Cn)
	out += ""

	//tb = table.DefaultSimpleTable(list)
	tb := _recordTable(records, evs, num)
	out += tb.String()

	return message.NewOutMessage(out), nil
}

func _recordTable(records []result.Record, evs []event.Event, num int) *table.Table {

	evMap := make(map[string]event.Event)

	for _, ev := range evs {
		evMap[ev.ID] = ev
	}

	var tbsMap = make(map[int]tableRecord)
	for idx, record := range records {
		tb := tableRecord{
			Name:     record.UserName,
			Event:    evMap[record.EventId].Cn,
			ResultID: int(record.ResultId),
			Time:     record.ResultTime.Format("2006/01/02"),
			Idx:      idx,
		}

		if t, ok := tbsMap[int(record.ResultId)]; ok {
			tb = t
		}

		if record.Best != nil || record.Repeatedly != nil {
			tb.Single = record.ResultString
		}
		if record.Average != nil {
			tb.Avg = record.ResultString
		}
		tbsMap[int(record.ResultId)] = tb
	}

	var list []tableRecord
	for _, val := range tbsMap {
		list = append(list, val)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Idx > list[j].Idx
	})

	for idx := range list {
		if list[idx].Single == "" {
			list[idx].Single = "-"
		}
		if list[idx].Avg == "" {
			list[idx].Avg = "-"
		}
		list[idx].Rank = idx + 1
	}

	if len(list) > num {
		list = list[:num]
	}

	tb, _ := table.SimpleTable(list, &table.Option{
		ExpendID: false,
		Align:    table.AlignLeft,
		Contour:  table.EmptyContour,
	})
	return tb
}
