package utils

import (
	"fmt"
	"strconv"
	"time"
)

func SecondTimeFormat(seconds float64, mbf bool) string {
	intSeconds := int64(seconds)
	decimalSeconds := int64(seconds*100) % 100
	duration := time.Duration(intSeconds) * time.Second

	hours := int64(duration.Hours())
	minutes := int64(duration.Minutes()) % 60
	secondsInt := int64(duration.Seconds()) % 60

	mmSecondsStr := fmt.Sprintf(".%02d", decimalSeconds)
	if decimalSeconds == 0 && (duration >= time.Hour || mbf) {
		mmSecondsStr = ""
	}
	//fmt.Println(fmt.Sprintf("%d:%02d:%02d%s", hours, minutes, secondsInt, mmSecondsStr))
	//return strings.TrimLeft(fmt.Sprintf("%d:%02d:%02d%s", hours, minutes, secondsInt, mmSecondsStr), "0:")
	if duration < time.Minute {
		return fmt.Sprintf("%d%s", secondsInt, mmSecondsStr)
	}
	if duration < time.Hour {
		return fmt.Sprintf("%d:%02d%s", minutes, secondsInt, mmSecondsStr)
	}
	return fmt.Sprintf("%d:%02d:%02d%s", hours, minutes, secondsInt, mmSecondsStr)
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
		// https://www.worldcubeassociation.org/export/results
		//difference    = 99 - DD
		//timeInSeconds = TTTTT (99999 means unknown)
		//missed        = MM
		//solved        = difference + missed
		//attempted     = solved + missed
		strIn := strconv.Itoa(in)
		diff, _ := strconv.Atoi(strIn[:2])
		miss, _ := strconv.Atoi(strIn[len(strIn)-2:])
		seconds, _ := strconv.Atoi(strIn[3 : len(strIn)-2])
		if seconds == 99999 {
			return "unknown"
		}
		formattedTime := SecondTimeFormat(float64(seconds), true)
		solved := 99 - diff + miss
		attempted := solved + miss
		return fmt.Sprintf("%d/%d %s", solved, attempted, formattedTime)
	}
	return SecondTimeFormat(float64(in)/100.0, false)
}
