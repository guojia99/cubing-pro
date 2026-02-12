package algs

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetAlgorithms(t *testing.T) {

	if err := Init("/home/guojia/worker/code/cube/cubing-pro/build/Alg-Trainers"); err != nil {
		t.Fatal(err)
	}

	out, _ := json.MarshalIndent(GetAlgorithms(), "", "    ")
	_ = os.WriteFile("algs.json", out, 0644)
}
