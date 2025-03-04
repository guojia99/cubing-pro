package crawler

import (
	"fmt"
	"testing"
)

func Test_getBaseCompetitionUrls(t *testing.T) {
	d := getBaseCompetitionUrls(2024)

	for _, v := range d {
		fmt.Println(v)
	}
	fmt.Println(len(d))
}

func Test_getAllCompetitionUrls(t *testing.T) {
	d := getAllCompetitionUrls()
	for _, v := range d {
		fmt.Println(v)
	}
	fmt.Println(len(d))
}

func Test_getPage(t *testing.T) {
	u := "https://cubing.com/competition/FMC-Cubing-China-2024"

	out, tt, _ := getPage(u)
	fmt.Println(out, tt)

	u = "https://cubing.com/competition/FMC-Cubing-China-2025"

	out, tt, _ = getPage(u)
	fmt.Println(out, tt)
}

func Test_checkAllCompetition(t *testing.T) {
	fd := CheckAllCubingCompetition()
	fmt.Println(fd)
}

func Test_getAllProbablyUrl(t *testing.T) {
	a, b := getAllProbablyUrl()
	fmt.Println(a, b)
	fmt.Println(len(a))
	fmt.Println(len(b))
}
