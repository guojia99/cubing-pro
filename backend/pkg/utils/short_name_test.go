package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestGetSName(t *testing.T) {

	names := []string{
		"訡殅緣未尽",
		"songphyxu",
		"郭嘉123",
		"1234567",
		"2016YEXI01",
		"Cuber奎",
		"Justkidding",
		"6223",
	}

	for _, name := range names {
		got := GetSName(time.Now(), name)
		fmt.Printf("--- %s === %s\n", name, got)
	}
}
