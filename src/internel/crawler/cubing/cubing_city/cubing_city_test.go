package cubing_city

import (
	"fmt"
	"testing"
)

func TestGetCubingCityList(t *testing.T) {
	city, _ := GetCubingCityListAndOldKey(2015, 2025)
	fmt.Println(city)
	fmt.Println(len(city))
}
