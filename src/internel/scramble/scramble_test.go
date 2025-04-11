package scramble

import (
	"fmt"
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
)

func Test_scramble_Scramble(t *testing.T) {
	s := &scramble{
		tNoodleEndpoint: "http://localhost:2014",
		scrambleType:    scrambleTypeRustTwisty,
	}

	t.Run("fto", func(t *testing.T) {
		data, err := s.ScrambleWithComp(event.Event{
			AutoScrambleKey: "FTO",
			BaseRouteType:   event.RouteType5RoundsAvgHT,
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("data: %v", data)
	})

	t.Run("333,444,555,666,777", func(t *testing.T) {
		data, err := s.ScrambleWithComp(event.Event{
			ScrambleValue: "333,444,555,666,777",
			BaseRouteType: event.RouteType1rounds,
		})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(data)
		for _, d := range data {
			t.Logf("data: %v", d)
		}
	})
}
