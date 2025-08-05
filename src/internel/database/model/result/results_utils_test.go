package result

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

func Test_sortRepeatedly(t *testing.T) {
	type args struct {
		in []repeatedly
	}
	tests := []struct {
		name string
		args args
		want []repeatedly
	}{
		{
			name: "正常排序",
			args: args{
				in: []repeatedly{
					{
						Reduction: 10,
						Try:       10,
						Time:      100,
					},
					{
						Reduction: 20,
						Try:       20,
						Time:      100,
					},
				},
			},
			want: []repeatedly{
				{
					Reduction: 20,
					Try:       20,
					Time:      100,
				},
				{
					Reduction: 10,
					Try:       10,
					Time:      100,
				},
			},
		},
		{
			name: "相等",
			args: args{
				in: []repeatedly{
					{
						Reduction: 10,
						Try:       10,
						Time:      100,
					},
					{
						Reduction: 10,
						Try:       10,
						Time:      100,
					},
				},
			},
			want: []repeatedly{
				{
					Reduction: 10,
					Try:       10,
					Time:      100,
				},
				{
					Reduction: 10,
					Try:       10,
					Time:      100,
				},
			},
		},
		{
			name: "依据时间排序",
			args: args{
				in: []repeatedly{
					{
						Reduction: 10,
						Try:       10,
						Time:      100,
					},
					{
						Reduction: 10,
						Try:       10,
						Time:      20,
					},
				},
			},
			want: []repeatedly{
				{
					Reduction: 10,
					Try:       10,
					Time:      20,
				},
				{
					Reduction: 10,
					Try:       10,
					Time:      100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := sortRepeatedly(tt.args.in); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("sortRepeatedly() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestResults_updateBestAndAvg(t *testing.T) {
	type fields struct {
		EventRoute event.RouteType
		Result     []float64
	}
	tests := []struct {
		name     string
		fields   fields
		wantBest float64
		wantAvg  float64
	}{
		{
			name: "RouteType1rounds",
			fields: fields{
				EventRoute: event.RouteType1rounds,
				Result:     []float64{1},
			},
			wantBest: 1,
			wantAvg:  DNF,
		},
		{
			name: "RouteType1rounds_DNF",
			fields: fields{
				EventRoute: event.RouteType1rounds,
				Result:     []float64{DNF},
			},
			wantBest: DNF,
			wantAvg:  DNF,
		},
		{
			name: "RouteType3roundsBest",
			fields: fields{
				EventRoute: event.RouteType3roundsBest,
				Result:     []float64{1, 2, 3},
			},
			wantBest: 1,
			wantAvg:  2,
		},
		{
			name: "RouteType3roundsBest_DNF",
			fields: fields{
				EventRoute: event.RouteType3roundsBest,
				Result:     []float64{1, 2, DNF},
			},
			wantBest: 1,
			wantAvg:  DNF,
		},
		{
			name: "RouteType3roundsAvg_DNF",
			fields: fields{
				EventRoute: event.RouteType3roundsAvg,
				Result:     []float64{3, 2, DNF},
			},
			wantBest: 2,
			wantAvg:  DNF,
		},
		{
			name: "RouteType3roundsAvg",
			fields: fields{
				EventRoute: event.RouteType3roundsAvg,
				Result:     []float64{2, 1, 3},
			},
			wantBest: 1,
			wantAvg:  2,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				c := &Results{
					EventRoute: tt.fields.EventRoute,
					Result:     tt.fields.Result,
				}
				_ = c.updateBestAndAvg()
				if c.Best != tt.wantBest {
					t.Errorf("updateBestAndAvg() best = %v, wantBest %v", c.Best, tt.wantBest)
				}

				if c.Average != tt.wantAvg {
					t.Errorf("updateBestAndAvg() Average = %v, wantAvg %v", c.Average, tt.wantAvg)
				}
			},
		)
	}
}

func Test_getBestAndAvg(t *testing.T) {
	tests := []struct {
		name     string
		results  []float64
		routeMap event.RouteMap
		wantBest float64
		wantAvg  float64
	}{
		{
			name:     "HeadToTailNum_OK",
			results:  []float64{3, 2, 1, 4, 5},
			routeMap: event.RouteType5RoundsAvgHT.RouteMap(),
			wantBest: 1,
			wantAvg:  3,
		},
		{
			name:     "HeadToTailNum2_DNF1",
			results:  []float64{3, 2, DNF, 4, 5},
			routeMap: event.RouteType5RoundsAvgHT.RouteMap(),
			wantBest: 2,
			wantAvg:  4.0,
		},
		{
			name:     "HeadToTailNum2_DNF2",
			results:  []float64{3, 2, DNF, 4, DNF},
			routeMap: event.RouteType5RoundsAvgHT.RouteMap(),
			wantBest: 2,
			wantAvg:  DNF,
		},
		{
			name:     "HeadToTailNum2_DNF3",
			results:  []float64{DNF, DNF, DNF, 5, 1},
			routeMap: event.RouteType5RoundsAvgHT.RouteMap(),
			wantBest: 1,
			wantAvg:  DNF,
		},
		{
			name:     "HeadToTailNum2_DNF4",
			results:  []float64{DNF, DNF, DNF, 5, DNF},
			routeMap: event.RouteType5RoundsAvgHT.RouteMap(),
			wantBest: 5,
			wantAvg:  DNF,
		},
		{
			name:     "HeadToTailNum2_DNF5",
			results:  []float64{DNF, DNF, DNF, DNF, DNF},
			routeMap: event.RouteType5RoundsAvgHT.RouteMap(),
			wantBest: DNF,
			wantAvg:  DNF,
		},
		{
			name:     "HeadToTailNum2_DNF_DNF",
			results:  []float64{DNS, DNS, DNF, DNF, DNF},
			routeMap: event.RouteType5RoundsAvgHT.RouteMap(),
			wantBest: DNF,
			wantAvg:  DNF,
		},
		{
			name:     "RouteType1rounds_OK",
			results:  []float64{1},
			routeMap: event.RouteType1rounds.RouteMap(),
			wantBest: 1,
			wantAvg:  DNF,
		},
		{
			name:     "RouteType1rounds_DNF",
			results:  []float64{DNS},
			routeMap: event.RouteType1rounds.RouteMap(),
			wantBest: DNF,
			wantAvg:  DNF,
		},
		{
			name:     "RouteType3roundsBest_OK",
			results:  []float64{1, 2, 3},
			routeMap: event.RouteType3roundsBest.RouteMap(),
			wantBest: 1,
			wantAvg:  2,
		},
		{
			name:     "RouteType3roundsBest_DNF",
			results:  []float64{2, 3, DNF},
			routeMap: event.RouteType3roundsBest.RouteMap(),
			wantBest: 2,
			wantAvg:  DNF,
		},
		{
			name:     "RouteType5roundsAvg_DNF",
			results:  []float64{1, 2, DNF, 4, 5},
			routeMap: event.RouteType5roundsAvg.RouteMap(),
			wantBest: 1,
			wantAvg:  DNF,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				gotBest, gotAvg := GetBestAndAvg(tt.results, tt.routeMap)
				if gotBest != tt.wantBest {
					t.Errorf("GetBestAndAvg() gotBest = %v, want %v", gotBest, tt.wantBest)
				}
				if gotAvg != tt.wantAvg {
					t.Errorf("GetBestAndAvg() gotAvg = %v, want %v", gotAvg, tt.wantAvg)
				}
			},
		)
	}
}

func TestUpdateOrgResult(t *testing.T) {
	type args struct {
		in           []float64
		eventRoute   event.RouteType
		cutoff       float64
		cutoffNumber int
		timeLimit    float64
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{
			name: "cut",
			args: args{
				in:           []float64{100, 200, DNF, 200, 300},
				eventRoute:   event.RouteType5RoundsAvgHT,
				cutoff:       150,
				cutoffNumber: 2,
				timeLimit:    250,
			},
			want: []float64{100, 200, DNF, 200, DNT},
		},
		{
			name: "limit_time",
			args: args{
				in:           []float64{1000, 1000, 2000, 2000, DNF},
				eventRoute:   event.RouteType5RoundsAvgHT,
				cutoff:       0,
				cutoffNumber: 0,
				timeLimit:    100,
			},
			want: []float64{DNT, DNT, DNT, DNT, DNF},
		},
		{
			name: "cut2_sq1_in_GuangzhouSpecial2024",
			args: args{
				in:           []float64{25.91, 27.05, 10, 10, 10},
				eventRoute:   event.RouteType5RoundsAvgHT,
				cutoff:       25,
				cutoffNumber: 2,
				timeLimit:    30,
			},
			want: []float64{25.91, 27.05, DNP, DNP, DNP},
		},
		{
			name: "add——DNS",
			args: args{
				in:           []float64{25},
				eventRoute:   event.RouteType5RoundsAvgHT,
				cutoff:       30,
				cutoffNumber: 2,
				timeLimit:    30,
			},
			want: []float64{25, DNS, DNS, DNS, DNS},
		},
		{
			name: "add--DNS-rep",
			args: args{
				in:           []float64{},
				eventRoute:   event.RouteTypeRepeatedly,
				cutoff:       0,
				cutoffNumber: 0,
				timeLimit:    0,
			},
			want: []float64{0, 0, DNS},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := UpdateOrgResult(tt.args.in, tt.args.eventRoute, tt.args.cutoff, tt.args.cutoffNumber, tt.args.timeLimit); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UpdateOrgResult() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestTimeParser(t *testing.T) {
	type args struct {
		in float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "max_2h",
			args: args{
				in: 9612,
			},
			want: "2:40:12",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeParserF2S(tt.args.in); got != tt.want {
				t.Errorf("TimeParser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeParserS2F(t *testing.T) {
	fmt.Println(TimeParserS2F("6.30"))
}
