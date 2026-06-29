package scramble

import (
	"math/rand"
)

var (
	FTOScrambleKey = [2][]string{
		{"R", "B", "L", "D", "F", "U"},
		{"R", "B", "L", "D", "F", "U", "BR", "BL"},
	}

	Cube444ScrambleKey = [2][]string{
		{"R", "B", "L", "D", "F", "U"},
		{"R", "B", "L", "D", "F", "U", "Fw", "Rw", "Uw"},
	}

	FTOAxisMap = map[string]string{
		"R": "RL",
		"L": "RL",

		"F": "FB",
		"B": "FB",

		"U": "UD",
		"D": "UD",

		"BR": "BRBL",
		"BL": "BRBL",
	}

	Cube444AxisMap = map[string]string{
		"R":  "RL",
		"L":  "RL",
		"Rw": "RL",

		"F":  "FB",
		"B":  "FB",
		"Fw": "FB",

		"U":  "UD",
		"D":  "UD",
		"Uw": "UD",
	}
)

func (s *scramble) autoScramble(
	keys [2][]string,
	axisMap map[string]string,
	min, max int,
	group int,
) []string {

	out := make([]string, 0, group)

	for i := 0; i < group; i++ {
		moveCount := rand.Intn(max-min+1) + min

		var (
			result    string
			lastMove  string
			lastMoves []string
		)

		for len(lastMoves) < moveCount {
			list := keys[0]
			if len(lastMoves) > min/2 {
				list = keys[1]
			}

			move := list[rand.Intn(len(list))]

			// 禁止连续同一个面
			if move == lastMove {
				continue
			}

			// 禁止连续三步同轴，例如：
			// U D U
			// R L R
			// R L Rw
			if len(lastMoves) >= 2 {
				axis := axisMap[move]

				lastAxis := axisMap[lastMoves[len(lastMoves)-1]]
				prevAxis := axisMap[lastMoves[len(lastMoves)-2]]

				if axis == lastAxis && axis == prevAxis {
					continue
				}
			}

			result += move

			if rand.Intn(2) == 1 {
				result += "'"
			}

			result += " "

			lastMove = move
			lastMoves = append(lastMoves, move)
		}

		out = append(out, result)
	}

	return out
}
