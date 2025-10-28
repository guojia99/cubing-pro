package wca_api

import (
	"fmt"
	"testing"
	"time"
)

func Test_getWcaSeniors(t *testing.T) {
	data, err := getWcaSeniors()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}

func Test_getSeniorsPerson(t *testing.T) {
	t.Run("2024GESH01", func(t *testing.T) {
		out, err := GetSeniorsPerson("2024GESH01")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(out)
	})
	t.Run("2017CHEL02", func(t *testing.T) {
		out, err := GetSeniorsPerson("2017CHEL02")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(out)
	})

}

func TestGetSeniorsWithEventsAndGroup(t *testing.T) {
	time.Sleep(time.Second * 5)
	bs, out, err := GetSeniorsWithEventsAndGroup(40, []string{"333mbf"})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", bs.Single["333mbf"])
	fmt.Printf("%+v\n", out)
}
