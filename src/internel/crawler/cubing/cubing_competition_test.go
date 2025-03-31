package cubing

import (
	"fmt"
	"strings"
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

	out, tt, _ := getPage("x", u)
	fmt.Println(out, tt)

	u = "https://cubing.com/competition/FMC-Cubing-China-2025"

	out, tt, _ = getPage("x", u)
	fmt.Println(out, tt)
}

func Test_checkAllCompetition(t *testing.T) {
	fd := CheckAllCubingCompetition()
	fmt.Println(fd)
}

func Test_getAllProbablyUrl(t *testing.T) {
	a, _ := getAllProbablyUrl()
	//fmt.Println(a)
	fmt.Println(a)

	for key, v := range a {
		if strings.Contains(key, "Shenyang") {
			fmt.Printf("contains => `%s`, `%s`\n", key, v)
		}
	}
	fmt.Println(a["Shenyang-Spring-2025"])

	//fmt.Println(len(a))
	//fmt.Println(len(b))
	//for k, v := range a {
	//	if k == "Shenyang" {
	//		fmt.Println(k, v)
	//	}
	//}
	//fmt.Println(b)
}

func Test_getPage1(t *testing.T) {
	out, _, _ := getPage("xxx", "https://cubing.com/competition/Please-Be-Quiet-Shanghai-2025")
	t.Logf("%+v", out)
}
