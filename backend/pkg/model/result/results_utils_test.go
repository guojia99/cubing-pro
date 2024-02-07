package result

import (
	"reflect"
	"testing"
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
		EventRoute RouteType
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
				EventRoute: RouteType1rounds,
				Result:     []float64{1},
			},
			wantBest: 1,
			wantAvg:  DNF,
		},
		{
			name: "RouteType1rounds_DNF",
			fields: fields{
				EventRoute: RouteType1rounds,
				Result:     []float64{DNF},
			},
			wantBest: DNF,
			wantAvg:  DNF,
		},
		{
			name: "RouteType3roundsBest",
			fields: fields{
				EventRoute: RouteType3roundsBest,
				Result:     []float64{1, 2, 3},
			},
			wantBest: 1,
			wantAvg:  2,
		},
		{
			name: "RouteType3roundsBest_DNF",
			fields: fields{
				EventRoute: RouteType3roundsBest,
				Result:     []float64{1, 2, DNF},
			},
			wantBest: 1,
			wantAvg:  DNF,
		},
		{
			name: "RouteType3roundsAvg_DNF",
			fields: fields{
				EventRoute: RouteType3roundsAvg,
				Result:     []float64{3, 2, DNF},
			},
			wantBest: 2,
			wantAvg:  DNF,
		},
		{
			name: "RouteType3roundsAvg",
			fields: fields{
				EventRoute: RouteType3roundsAvg,
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
