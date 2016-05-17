package algorithms

import (
	"math"
	"sort"
)

type BeatAnalysis struct {
	Unstable         bool  `json:"unstable,omitempty"`
	LowerBound       int64 `json:"lower"`
	UpperBound       int64 `json:"upper"`
	PercentageWithin int   `json:"percentage"`
}

type int64arr []int64

func (a int64arr) Len() int {
	return len(a)
}
func (a int64arr) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a int64arr) Less(i, j int) bool {
	return a[i] < a[j]
}

func AnalyzeBeats(beats []int64) *BeatAnalysis {
	if len(beats) < 10 {
		return nil
	}

	var analysis BeatAnalysis
	sorted := beats[:]
	sort.Sort(int64arr(sorted))
	analysis.LowerBound = int64(math.Floor(float64(sorted[int(float64(len(sorted))*0.1)])/1000 + 0.5))
	analysis.UpperBound = int64(math.Floor(float64(sorted[int(float64(len(sorted))*0.9)])/1000 + 0.5))
	analysis.PercentageWithin = 90

	if analysis.UpperBound >= 5*analysis.LowerBound {
		analysis.Unstable = true
	}

	return &analysis
}
