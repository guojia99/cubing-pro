package utils

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func SecondTimeFormat(seconds float64, mbf bool) string {
	if seconds <= 59.99 && !mbf {
		return fmt.Sprintf("%.2f", seconds)
	}

	// 分离整数秒和小数秒
	intSeconds := int64(seconds)
	totalSeconds := intSeconds
	decimalSeconds := int64(math.Round((seconds - float64(intSeconds)) * 100))

	// 计算时分秒
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secondsInt := totalSeconds % 60

	mmSecondsStr := fmt.Sprintf(".%02d", decimalSeconds)
	if decimalSeconds == 0 && (totalSeconds >= 3600 || mbf) {
		mmSecondsStr = ""
	}

	if totalSeconds < 60 {
		return fmt.Sprintf("%d%s", secondsInt, mmSecondsStr)
	}
	if totalSeconds < 3600 {
		return fmt.Sprintf("%d:%02d%s", minutes, secondsInt, mmSecondsStr)
	}
	return fmt.Sprintf("%d:%02d:%02d%s", hours, minutes, secondsInt, mmSecondsStr)
}

func Get333MBFResult(in int) (solved, attempted int, seconds int, formattedTime string) {
	// https://www.worldcubeassociation.org/export/results
	//difference    = 99 - DD
	//timeInSeconds = TTTTT (99999 means unknown)
	//missed        = MM
	//solved        = difference + missed
	//attempted     = solved + missed
	strIn := strconv.Itoa(in)
	diff, _ := strconv.Atoi(strIn[:2])
	miss, _ := strconv.Atoi(strIn[len(strIn)-2:])
	seconds, _ = strconv.Atoi(strIn[3 : len(strIn)-2])
	//if seconds == 99999 {
	//	return "unknown"
	//}

	formattedTime = SecondTimeFormat(float64(seconds), true)
	solved = 99 - diff + miss
	attempted = solved + miss
	return
}

func ResultsTimeFormat(in int, event string) string {
	switch in {
	case -1:
		return "DNF"
	case -2:
		return "DNS"
		// todo other particular result
	default:
	}

	switch event {
	case "333fm":
		if in > 1000 {
			return fmt.Sprintf("%.2f", float64(in)/100.0)
		}
		return fmt.Sprintf("%d", in)
	case "333mbf":
		solved, attempted, _, formattedTime := Get333MBFResult(in)
		return fmt.Sprintf("%d/%d %s", solved, attempted, formattedTime)
	}
	return SecondTimeFormat(float64(in)/100.0, false)
}

func IsBestResult(event string, a1, a2 int) bool {
	// DNF
	if a1 < 0 && a2 < 0 {
		return true
	}
	if a1 < 0 && a2 > 0 {
		return false
	}
	if a2 < 0 && a1 > 0 {
		return true
	}

	switch event {
	case "333mbf":
		a1Solved, a1Attempted, a1Seconds, _ := Get333MBFResult(a1)
		a2Solved, a2Attempted, a2Seconds, _ := Get333MBFResult(a2)

		// 先比分数
		a1Res := a1Solved - (a1Attempted - a1Solved)
		a2Res := a2Solved - (a2Attempted - a2Solved)

		if a1Res != a2Res {
			return a1Res > a2Res
		}
		return a1Seconds <= a2Seconds
	default:
		return a1 <= a2
	}
}

func ParserTimeToSeconds(t string) float64 {
	// 解析纯秒数格式
	if regexp.MustCompile(`^\d+(\.\d+)?$`).MatchString(t) {
		seconds, _ := strconv.ParseFloat(t, 64)
		return seconds
	}

	// 解析分+秒格式
	if regexp.MustCompile(`^\d{1,3}:\d{1,3}(\.\d+)?$`).MatchString(t) {
		parts := strings.Split(t, ":")
		minutes, _ := strconv.ParseFloat(parts[0], 64)
		seconds, _ := strconv.ParseFloat(parts[1], 64)
		return minutes*60 + seconds
	}

	// 解析时+分+秒格式
	if regexp.MustCompile(`^\d{1,3}:\d{1,3}:\d{1,3}(\.\d+)?$`).MatchString(t) {
		parts := strings.Split(t, ":")
		hours, _ := strconv.ParseFloat(parts[0], 64)
		minutes, _ := strconv.ParseFloat(parts[1], 64)
		seconds, _ := strconv.ParseFloat(parts[2], 64)
		return hours*3600 + minutes*60 + seconds
	}

	return -1
}

func WCAResultIntToSeconds(wcaResult int, event string) float64 {
	switch event {
	case "333mbf":
		return 0
	case "333fm":
		return float64(wcaResult)
	default:
	}

	return float64(wcaResult) / 100.0
}
