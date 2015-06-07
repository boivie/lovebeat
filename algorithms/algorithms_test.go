package algorithms

import (
	"fmt"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func TestSimple(t *testing.T) {
	input := []float64{60000, 60000}
	ret := AutoAlg(input, 3)

	assertEqual(t, ret, 61000, "")
}

func TestRange(t *testing.T) {
	input := []float64{60000, 60000, 60000, 60000, 60000, 60000}
	ret := AutoAlg(input, 3)

	assertEqual(t, ret, 61000, "")
}
