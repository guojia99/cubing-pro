package cubing_city

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/utils"
)

func TestOtherCitys(t *testing.T) {
	out, _ := json.Marshal(utils.RemoveDuplicates(zheJiangAndJiangSuCitys))
	os.WriteFile("test3.json", out, 0644)
}
