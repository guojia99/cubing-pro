package staticx

import (
	"fmt"
	"testing"
)

func TestStaticX_GetStatic(t *testing.T) {
	s := StaticX{}
	s.Init()
	//s.BaseData(false)
	data := s.GetStatic("China", 2024)
	fmt.Println(data)
	//data2 := s.GetStatic("Hong Kong", 2024)
	//fmt.Println(data2)
	//data3 := s.GetStatic("TaiWan", 2024)
	//fmt.Println(data3)
	//data4 := s.GetStatic("Macau", 2024)
	//fmt.Println(data4)
}
