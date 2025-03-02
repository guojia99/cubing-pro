package scramble

/*
#cgo LDFLAGS: -L./ -lrust_scramble -ldl -lm
#include "rust_scramble.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"time"
)

func cube333Scramble() string {
	o := C.cube333_scramble()
	output := C.GoString(o)
	return output
}
func cube222Scramble() string {
	o := C.cube222_scramble()
	output := C.GoString(o)
	return output
}
func cube444Scramble() string {
	o := C.cube444_scramble()
	output := C.GoString(o)
	return output
}
func cube555Scramble() string {
	o := C.cube555_scramble()
	output := C.GoString(o)
	return output
}
func cube666Scramble() string {
	o := C.cube666_scramble()
	output := C.GoString(o)
	return output
}
func cube777Scramble() string {
	o := C.cube777_scramble()
	output := C.GoString(o)
	return output
}
func cube333bfScramble() string {
	o := C.cube333bf_scramble()
	output := C.GoString(o)
	return output
}
func cube333fmScramble() string {
	o := C.cube333fm_scramble()
	output := C.GoString(o)
	return output
}
func cube333ohScramble() string {
	o := C.cube333oh_scramble()
	output := C.GoString(o)
	return output
}
func clockScramble() string {
	o := C.clock_scramble()
	output := C.GoString(o)
	return output
}
func megaminxScramble() string {
	o := C.megaminx_scramble()
	output := C.GoString(o)
	return output
}
func pyraminxScramble() string {
	o := C.pyraminx_scramble()
	output := C.GoString(o)
	return output
}
func skewbScramble() string {
	o := C.skewb_scramble()
	output := C.GoString(o)
	return output
}
func sq1Scramble() string {
	o := C.sq1_scramble()
	output := C.GoString(o)
	return output
}
func cube444bfScramble() string {
	o := C.cube444bf_scramble()
	output := C.GoString(o)
	return output
}
func cube555bfScramble() string {
	o := C.cube555bf_scramble()
	output := C.GoString(o)
	return output
}
func cube333ftScramble() string {
	o := C.cube333_scramble()
	output := C.GoString(o)
	return output
}

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

// Rust静态库本代码由狼(2007YUNQ01) 提供，为rust编写的打乱生成器。
func (s *scramble) rustScramble(cube string, nums int) ([]string, error) {
	var out []string
	for i := 0; i < nums; i++ {
		fn, ok := rustScrambleMp[cube]
		if !ok {
			return nil, errors.New("cube not found")
		}
		out = append(out, fn())
	}
	return out, nil
}

func (s *scramble) rustTestLongScramble() string {
	out := ""
	testFn := func(key string, fn func() string) {
		var times []time.Duration
		for i := 0; i < 5; i++ {
			start := time.Now()
			_ = fn()
			duration := time.Since(start)
			times = append(times, duration)
		}

		// 计算最大值、最小值和平均值
		var minS, maxS, sum time.Duration = times[0], times[0], 0
		for _, t := range times {
			if t < minS {
				minS = t
			}
			if t > maxS {
				maxS = t
			}
			sum += t
		}
		avg := sum / time.Duration(len(times))
		out += fmt.Sprintf("%s => 最小:%v;最大:%v;平均: %v\n", key, minS, maxS, avg)
	}

	for key, fn := range rustScrambleMp {
		testFn(key, fn)
	}
	return out
}
