package wca_api

import (
	"fmt"
	"testing"
)

func TestApiSearchPersons(t *testing.T) {
	ps, err := ApiSearchPersons("徐永浩")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ps)
}
