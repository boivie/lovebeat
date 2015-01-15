package service

import (
	"testing"
)

func testEq(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestMedianOdd(t *testing.T) {
	v := []int64{1, 3, 2, 5, 4}
	if median(v) != 3 {
		t.Errorf("Median was %d", median(v))
	}
}

func TestMedianEven(t *testing.T) {
	v := []int64{7, 5, 1, 3}
	if median(v) != 4 {
		t.Errorf("Median was %d", median(v))
	}
}

func TestCalcDiffs1(t *testing.T) {
	v := []int64{40, 50}
	u := calcDiffs(v)
	if !testEq(u, []int64{10}) {
		t.Errorf("Failed")
	}
}

func TestCalcDiffs2(t *testing.T) {
	v := []int64{4, 8, 15, 16}
	u := calcDiffs(v)
	if !testEq(u, []int64{4, 7, 1}) {
		t.Errorf("Failed")
	}
}

func TestCalcTimeouts1(t *testing.T) {
	v := []int64{4, 8, 15, 16}
	u := calcTimeout(v)
	if u != TIMEOUT_AUTO {
		t.Errorf("Failed")
	}
}

func TestCalcTimeouts2(t *testing.T) {
	v := []int64{10, 20, 30, 40, 50, 60}
	u := calcTimeout(v)
	if u != 15 {
		t.Errorf("Failed")
	}
}
