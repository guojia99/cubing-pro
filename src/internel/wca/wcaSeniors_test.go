package wca

import (
	"fmt"
	"testing"
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
