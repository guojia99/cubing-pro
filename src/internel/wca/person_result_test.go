package wca

import (
	"fmt"
	"testing"
)

func TestApiGetWCAResults(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		got, err := GetWCAPersonResult("2017XUYO01")

		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(got.String())
	})

	t.Run("test_2", func(t *testing.T) {
		got, err := GetWCAPersonResult("2013LINK01")

		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(got.String())
	})

	t.Run("test_cn", func(t *testing.T) {
		got, err := GetWCAPersonResult("徐永浩")

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(got)
	})
}
