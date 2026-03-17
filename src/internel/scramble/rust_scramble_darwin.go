//go:build darwin

package scramble

import (
	"fmt"
	"strings"
)

// 在 macOS 上，这些函数返回空字符串或模拟数据，不调用 Rust
func cube333Scramble() string   { return "" }
func cube222Scramble() string   { return "" }
func cube444Scramble() string   { return "" }
func cube555Scramble() string   { return "" }
func cube666Scramble() string   { return "" }
func cube777Scramble() string   { return "" }
func cube333bfScramble() string { return "" }
func cube333fmScramble() string { return "" }
func cube333ohScramble() string { return "" }
func clockScramble() string     { return "" }
func megaminxScramble() string  { return "" }
func pyraminxScramble() string  { return "" }
func skewbScramble() string     { return "" }
func sq1Scramble() string       { return "" }
func cube444bfScramble() string { return "" }
func cube555bfScramble() string { return "" }
func cube333ftScramble() string { return "" }

var rustScrambleMp = map[string]func() string{
	"333":    cube333Scramble,
	"222":    cube222Scramble,
	"555":    cube555Scramble,
	"666":    cube666Scramble,
	"777":    cube777Scramble,
	"333bf":  cube333bfScramble,
	"333fm":  cube333fmScramble,
	"333oh":  cube333ohScramble,
	"clock":  clockScramble,
	"minx":   megaminxScramble,
	"pyram":  pyraminxScramble,
	"skewb":  skewbScramble,
	"sq1":    sq1Scramble,
	"555bf":  cube555bfScramble,
	"333ft":  cube333ftScramble,
	"333mbf": cube333bfScramble,
	"444":    cube444Scramble,
	"444bf":  cube444bfScramble,
}

// rustScramble 在 macOS 上返回错误提示或空列表，避免崩溃
func (s *scramble) rustScramble(cube string, nums int) ([]string, error) {
	// 选项 A: 直接返回错误，告诉用户 macOS 不支持
	// return nil, errors.New("rust scramble not supported on macOS (disabled for build)")

	// 选项 B: 返回空结果但不报错 (根据你的需求 "保留可用的接口（无效接口即可）")
	// 这里生成 nums 个空字符串或者提示语
	out := make([]string, nums)
	for i := 0; i < nums; i++ {
		out[i] = "[macOS-disabled]"
	}
	return out, nil
}

func (s *scramble) rustTestLongScramble() string {
	var sb strings.Builder
	sb.WriteString("Rust scramble benchmark skipped on macOS.\n")

	// 模拟跑一下循环，但不实际执行耗时操作，仅展示接口可用
	for key := range rustScrambleMp {
		sb.WriteString(fmt.Sprintf("%s => Skipped (No Rust backend)\n", key))
	}
	return sb.String()
}
