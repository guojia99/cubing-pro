package staticx

import (
	"fmt"
	"os"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func TestStaticX_DNFEvents(t *testing.T) {
	s := StaticX{}
	s.Init()
	//s.BaseData(false)
	data := s.DNFEvents("China")
	d, _ := jsoniter.MarshalToString(data)
	fmt.Println(d)

	_ = os.WriteFile("China.json", []byte(d), 0644)
}
