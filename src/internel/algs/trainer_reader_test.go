package algs

import (
	"fmt"
	"testing"
)

func TestReadTrainerFiles(t *testing.T) {
	data, err := ReadTrainerFiles("/home/guojia/worker/code/cube/cubing-pro/build/Alg-Trainers/2x2-EG-Trainer")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(data.SetKeys)
}
