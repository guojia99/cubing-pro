package scramble

import (
	"fmt"
	"testing"
)

func Test_scramble_rustScramble(t *testing.T) {
	for key, fn := range rustScrambleMp {
		t.Run(key, func(t *testing.T) {
			out := fn()
			fmt.Println(out)
		})
	}
}
