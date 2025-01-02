package utils

import (
	"fmt"
	"testing"
)

func TestResultsTimeFormat(t *testing.T) {
	type args struct {
		in    int
		event string
	}
	tests := []struct {
		args args
		want string
	}{
		{args: args{in: 23, event: "333"}, want: "0.23"},
		{args: args{in: 1023, event: "333"}, want: "10.23"},
		{args: args{in: 123, event: "333"}, want: "1.23"},
		{args: args{in: 6001, event: "333"}, want: "1:00.01"},
		{args: args{in: 6101, event: "333"}, want: "1:01.01"},
		{args: args{in: 6200, event: "333"}, want: "1:02.00"},
		{args: args{in: 360000, event: "333"}, want: "1:00:00"},
		{args: args{in: 359999, event: "333"}, want: "59:59.99"},

		// fm
		{args: args{in: 20, event: "333fm"}, want: "20"},
		{args: args{in: 2067, event: "333fm"}, want: "20.67"},

		// mbf  results
		{args: args{in: 990196902, event: "333mbf"}, want: "2/4 32:49"},     // 2005AKKE01
		{args: args{in: 890331901, event: "333mbf"}, want: "11/12 55:19"},   // 2005AKKE01
		{args: args{in: 690360208, event: "333mbf"}, want: "38/46 1:00:02"}, // 2015ALEK01
		{args: args{in: 950360014, event: "333mbf"}, want: "18/32 1:00:00"}, // 2015ALEK01

		//{args: args{in: 832179317, event: "333mbo"}, want: "33/50 29:53"},

		// DNF DNS
		{args: args{in: -1, event: "333mbf"}, want: "DNF"},
		{args: args{in: -2, event: "333"}, want: "DNS"},
	}
	for _, tt := range tests {
		t.Run(
			fmt.Sprintf("%s-%d", tt.args.event, tt.args.in), func(t *testing.T) {
				if got := ResultsTimeFormat(tt.args.in, tt.args.event); got != tt.want {
					t.Errorf("ResultsTimeFormat() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
