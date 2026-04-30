package wca

import (
	"fmt"
	"testing"

	"github.com/guojia99/cubing-pro/src/wca/types"
)

func Test_wca_ResultProportionEstimation(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, err := w.ResultProportionEstimation(types.ResultProportionEstimationTypeBigCube, 10)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(out)
}

func Test_resultProportionEstimationData_synthetic(t *testing.T) {
	events := []string{"444", "555", "666"}
	data := map[string]map[string][]types.Result{
		"p1": {
			"444": {{Attempts: []int64{2000, 2010}}},
			"555": {{Attempts: []int64{4000, 4010}}},
			"666": {{Attempts: []int64{8000, 8010}}},
		},
		"p2": {
			"444": {{Attempts: []int64{3000}}},
			"555": {{Attempts: []int64{6000}}},
			"666": {{Attempts: []int64{12000}}},
		},
		"p3": {
			"444": {{Attempts: []int64{4000}}},
			"555": {{Attempts: []int64{8000}}},
			"666": {{Attempts: []int64{16000}}},
		},
	}
	out, err := resultProportionEstimationData(events, data)
	if err != nil {
		t.Fatal(err)
	}
	if out.SampleCount != 3 {
		t.Fatalf("SampleCount = %d, want 3", out.SampleCount)
	}
	if out.GlobalRatio["555"] < 1.9 || out.GlobalRatio["555"] > 2.1 {
		t.Fatalf("GlobalRatio[555] = %v, want ~2", out.GlobalRatio["555"])
	}
	if out.GlobalRatio["666"] < 3.9 || out.GlobalRatio["666"] > 4.1 {
		t.Fatalf("GlobalRatio[666] = %v, want ~4", out.GlobalRatio["666"])
	}
	if len(out.Segments) != 1 {
		t.Fatalf("n<5 应单段, got %d segments", len(out.Segments))
	}
	if len(out.CurveSamples) != 50 {
		t.Fatalf("CurveSamples len = %d", len(out.CurveSamples))
	}
}

func Test_getCompetitionIDYear(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "10JohrWilerWurfelfast2024",
			args: args{
				id: "10JohrWilerWurfelfast2024",
			},
			want:    2024,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCompetitionIDYear(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCompetitionIDYear() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCompetitionIDYear() got = %v, want %v", got, tt.want)
			}
		})
	}
}
