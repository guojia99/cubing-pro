package cubing

import (
	"fmt"
	"testing"
)

func TestGetAllWcaComps(t *testing.T) {
	fmt.Println(GetAllWcaComps())
}

func TestGetWcaInfo(t *testing.T) {
	t.Run("test-not", func(t *testing.T) {
		out := GetWcaInfo("xxxx-not")
		fmt.Println(out)
	})
}
