package scramble

import (
	"fmt"
	"testing"
	"time"
)

func Test_scramble_rustScramble(t *testing.T) {

	s := NewScramble(scrambleTypeRustTwisty, "").(*scramble)
	time.Sleep(10 * time.Second)

	t.Run("444", func(t *testing.T) {
		out, err := s.rustScramble("444", 5)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(out)
	})
}
