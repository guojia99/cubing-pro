package scramble

import (
	"math/rand"
)

var (
	FTOScrambleKey = [2][]string{
		{"R", "B", "L", "D", "F", "U"},             // 前半段
		{"R", "B", "L", "D", "F", "U", "BR", "BL"}, // 后半段
	}
)

func (s *scramble) AutoScramble(keys [2][]string, min, max int, group int) []string {
	var out []string

	for i := 0; i < group; i++ {
		data := ""
		curNum := rand.Intn(max-min+1) + min

		last := ""

		for j := 0; j < curNum; {
			list := keys[0]
			if j > min/2 {
				list = keys[1]
			}

			randomIndex := rand.Intn(len(list))
			randomKey := list[randomIndex]
			if randomKey == last {
				continue
			}
			data += randomKey
			if pr := rand.Intn(2) == 1; pr {
				data += "'"
			}
			data += " "
			last = randomKey
			j++
		}
		out = append(out, data)
	}
	return out
}
