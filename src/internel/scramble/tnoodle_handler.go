package scramble

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type scramble struct {
	url string
}

func (s scramble) get(typ string, nums int) (string, error) {
	url := fmt.Sprintf("%s/scramble/.txt?=%s*%d", s.url, typ, nums)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	v, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (s scramble) CubeScramble(cube string, nums int) ([]string, error) {
	key, err := TNoodleKey(cube)
	if err != nil {
		return nil, err
	}

	if nums >= 100 {
		nums = 100
	}

	data, err := s.get(key, nums)
	if err != nil {
		return nil, err
	}
	out := strings.Split(data, "\r\n")
	return out[:nums], err
}

/*
CE	基础公式	U' R' U L U' R U L'	[U' R' U,L]
EC	基础公式	L U' R' U L' U' R U	[L,U' R' U]
CM	基础公式	U' R' U L' U' R U L	[U' R' U,L']
MC	基础公式	L' U' R' U L U' R U	[L',U' R' U]
CQ	基础公式	U' R' U L2 U' R U L2	[U' R' U,L2]
QC	基础公式	L2 U' R' U L2 U' R U	[L2,U' R' U]
CH	基础公式	R' U L U' R U L' U'	[R',U L U']
HC	基础公式	U L U' R' U L' U' R	[U L U',R']
CR	基础公式	R2 U L U' R2 U L' U'	[R2,U L U']
RC	基础公式	U L U' R2 U L' U' R2	[U L U',R2]
CY	基础公式	R U L U' R' U L' U'	[R,U L U']
YC	基础公式	U L U' R U L' U' R'	[U L U',R]

*/
