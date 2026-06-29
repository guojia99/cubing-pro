package scramble

import (
	"testing"
)

func Test_scramble_autoScramble(t *testing.T) {

	s := &scramble{}

	t.Run("fto", func(t *testing.T) {
		for _, i := range s.autoScramble(FTOScrambleKey, FTOAxisMap, 25, 30, 10) {
			t.Logf("fto: %s\n", i)
		}
	})

	t.Run("444", func(t *testing.T) {
		for _, i := range s.autoScramble(Cube444ScrambleKey, Cube444AxisMap, 39, 49, 10) {
			t.Logf("444: %s\n", i)
		}
	})
}
