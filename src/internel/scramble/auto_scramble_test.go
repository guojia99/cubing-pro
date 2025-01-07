package scramble

import (
	"fmt"
	"testing"
)

func Test_scramble_AutoScramble(t *testing.T) {
	s := &scramble{}

	fmt.Println(s.AutoScramble(FTOScrambleKey, 25, 30, 1))
}
