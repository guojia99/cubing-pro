package tools

import (
	"fmt"
	"testing"

	"github.com/guojia99/cubing-pro/src/robot/types"
)

func TestTWca_Do(t1 *testing.T) {
	t := &TWca{}

	out, err := t.Do(types.InMessage{
		Message: "Wca 徐永浩",
	})
	if err != nil {
		t1.Fatal(err)
	}
	for _, msg := range out.Message {
		fmt.Println(msg)
	}
}
