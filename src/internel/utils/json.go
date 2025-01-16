package utils

import (
	"os"
)

func ReadJson(fileName string, v interface{}) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
	}
	return err
}
