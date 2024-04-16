package scramble

type Scramble interface {
	CubeScramble(cube string, nums int) ([]string, error)
}

func NewScramble(endpoint string) Scramble {
	return &scramble{
		url: endpoint,
	}
}
