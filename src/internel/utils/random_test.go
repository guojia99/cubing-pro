package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestGenerateRandomKey(t *testing.T) {
	ts := time.Now().UnixNano()

	fmt.Println(GenerateRandomKey(ts))
	fmt.Println(GenerateRandomKey(ts))
}
