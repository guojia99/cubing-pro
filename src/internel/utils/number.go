package utils

import (
	"math"
	"regexp"
	"strconv"
)

func GetNumbers(in string) []float64 {
	re := regexp.MustCompile("(-?\\d+)(\\.\\d+)?")
	numbers := re.FindAllString(in, -1)

	var out []float64
	for _, num := range numbers {
		f, err := strconv.ParseFloat(num, 64)
		if err == nil {
			out = append(out, f)
		}
	}
	return out
}

func GetNum(in string) float64 {
	numbers := GetNumbers(in)
	if len(numbers) > 0 {
		return numbers[0]
	}
	return math.NaN()
}
