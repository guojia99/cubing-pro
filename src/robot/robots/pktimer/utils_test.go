package pktimer

import (
	"testing"

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
			gotOut := getCurPackerMessage(tt.args.results)
			t.Logf("gotOut:%v", gotOut)
		})
	}
}
