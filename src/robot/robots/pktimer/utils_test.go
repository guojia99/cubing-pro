package pktimer

import (
	"testing"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	pktimerDB "github.com/guojia99/cubing-pro/src/internel/database/model/pktimer"
)

func Test_getCurPackerMessage(t *testing.T) {
	type args struct {
		results *pktimerDB.PkTimerResult
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "base",
			args: args{
				results: &pktimerDB.PkTimerResult{
					PkResults: pktimerDB.PkResults{
						Players: []pktimerDB.Player{
							{
								UserName: "嘉",
								Results:  []float64{10.00},
							},
							{
								UserName: "嘉2",
								Results:  []float64{10.01},
							},
						},
						CurCount: 1,
					},
				},
			},
		},
		{
			name: "exxpp",
			args: args{
				results: &pktimerDB.PkTimerResult{
					PkResults: pktimerDB.PkResults{
						Players: []pktimerDB.Player{
							{
								UserName: "嘉",
								Results:  []float64{10.00},
							},
							{
								UserName: "嘉2",
								Results:  []float64{10.00},
							},
						},
						CurCount: 1,
					},
				},
			},
		},
		{
			name: "not",
			args: args{
				results: &pktimerDB.PkTimerResult{
					PkResults: pktimerDB.PkResults{
						Players: []pktimerDB.Player{
							{
								UserName: "嘉",
								Results:  []float64{10.00},
							},
							{
								UserName: "嘉2",
								Results:  []float64{5.00},
							},
						},
						CurCount: 1,
					},
				},
			},
		},
		{
			name: "not",
			args: args{
				results: &pktimerDB.PkTimerResult{
					PkResults: pktimerDB.PkResults{
						Players: []pktimerDB.Player{
							{
								UserName: "嘉",
								Results:  []float64{10.00},
							},
							{
								UserName: "嘉2",
								Results:  []float64{10.05},
							},
						},
						CurCount: 1,
					},
				},
			},
		},
		{
			name: "not2",
			args: args{
				results: &pktimerDB.PkTimerResult{
					PkResults: pktimerDB.PkResults{
						Players: []pktimerDB.Player{
							{
								UserName: "嘉",
								Results:  []float64{10.00},
							},
							{
								UserName: "嘉2",
								Results:  []float64{10.05},
							},
							{
								UserName: "嘉23",
								Results:  []float64{10.07},
							},
						},
						CurCount: 1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut := getCurPackerMessage(tt.args.results, "cur")
			t.Logf("gotOut:%v", gotOut)
		})
	}
}

func Test_getAllPackerMessage(t *testing.T) {
	type args struct {
		results *pktimerDB.PkTimerResult
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "avg",
			args: args{
				results: &pktimerDB.PkTimerResult{
					PkResults: pktimerDB.PkResults{
						Players: []pktimerDB.Player{
							{
								UserName: "嘉",
								Average:  5.85,
							},
							{
								UserName: "乔治",
								Average:  5.81,
							},
						},
						CurCount: 1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut := getAllPackerMessage(tt.args.results)
			t.Logf("gotOut:%v", gotOut)
		})
	}
}

func Test_getAllPackerMessage1(t *testing.T) {
	results := &pktimerDB.PkTimerResult{
		PkResults: pktimerDB.PkResults{
			Players: []pktimerDB.Player{
				{
					UserName: "Clansey",
					Best:     16.25,
					Average:  18.77,
				},
				{
					UserName: "嘉",
					Best:     17.00,
					Average:  18.73,
				},
			},
			Event: event.Event{
				StringIDModel: basemodel.StringIDModel{
					ID: "333oh",
				},
				Cn:            "三单",
				BaseRouteType: 7,
			},
			Count:    5,
			CurCount: 5,
		},
		Eps: 0.10,
	}

	out := getAllPackerMessage(results)
	for _, v := range out {
		t.Logf("%v\n", v)
	}
}
