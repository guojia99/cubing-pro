package utils

import (
	"os"

	"github.com/2mf8/Better-Bot-Go/log"
)

func ReadJson(fileName string, v interface{}) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Errorf("Read json file %s error: %v", fileName, err)
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		log.Errorf("Unmarshal json file %s error: %v", fileName, err)
	}
	return err
}
