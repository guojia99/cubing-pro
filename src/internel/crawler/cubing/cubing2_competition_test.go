package cubing

import (
	"fmt"
	"testing"
)

func TestDCubingCompetition_GetNewCompetitions(t *testing.T) {
	c := NewDCubingCompetition()
	fmt.Println(c.GetNewCompetitions())
}
