package service

import (
	"math"
	"sort"
)

const (
	// Number of samples (diffs) we require to be able to
	// properly calculate an "auto" timeout
	AUTO_MIN_SAMPLES = 5
)

func calcTimeout(values []int64) int64 {
	diffs := calcDiffs(values)
	if len(diffs) < AUTO_MIN_SAMPLES {
		log.Debug("AUTO-TIMEOUT: Not enough samples to calculate")
		return TIMEOUT_AUTO
	}

	ret := int64(math.Ceil(float64(median(diffs)) * 1.5))
	log.Debug("AUTO-TIMEOUT: vale calculated as %d", ret)
	return ret
}

func calcDiffs(values []int64) []int64 {
	var p []int64
	for i := 1; i < len(values); i++ {
		if values[i-1] != 0 && values[i] != 0 {
			p = append(p, values[i]-values[i-1])
		}
	}
	return p
}

type int64arr []int64

func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

func median(numbers []int64) int64 {
	sort.Sort(int64arr(numbers))
	middle := len(numbers) / 2
	result := numbers[middle]
	if len(numbers)%2 == 0 {
		result = (result + numbers[middle-1]) / 2
	}
	return result
}
