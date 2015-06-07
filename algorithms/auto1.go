package algorithms

import (
	"math"
)

func unDef(f float64) bool {
	if math.IsNaN(f) {
		return true
	}
	if math.IsInf(f, 1) {
		return true
	}
	if math.IsInf(f, -1) {
		return true
	}
	return false
}

func Ewma(series []float64, com float64) []float64 {
	var cur float64
	var prev float64
	var oldw float64
	var adj float64
	N := len(series)
	ret := make([]float64, N)
	if N == 0 {
		return ret
	}
	oldw = com / (1 + com)
	adj = oldw
	ret[0] = series[0] / (1 + com)
	for i := 1; i < N; i++ {
		cur = series[i]
		prev = ret[i-1]
		if unDef(cur) {
			ret[i] = prev
		} else {
			if unDef(prev) {
				ret[i] = cur / (1 + com)
			} else {
				ret[i] = (com*prev + cur) / (1 + com)
			}
		}
	}
	for i := 0; i < N; i++ {
		cur = ret[i]
		if !math.IsNaN(cur) {
			ret[i] = ret[i] / (1. - adj)
			adj *= oldw
		} else {
			if i > 0 {
				ret[i] = ret[i-1]
			}
		}
	}
	return ret
}

func EwmStdLast(series []float64, com float64) float64 {
	m1st := Ewma(series, com)
	var series2 []float64
	for _, val := range series {
		series2 = append(series2, val*val)
	}
	m2nd := Ewma(series2, com)

	i := len(m1st) - 1
	t := m2nd[i] - math.Pow(m1st[i], 2)
	t *= (1.0 + 2.0*com) / (2.0 * com)
	if t < 0 {
		return 0
	}
	return math.Sqrt(t)
}

func copyLast(series []float64, N int) []float64 {
	var start = len(series) - N
	if start < 0 {
		start = 0
	}
	var end = len(series)
	return series[start:end]
}

func round(f float64) int {
	return int(math.Floor(f + .5))
}

func AutoAlg(series []float64, factor float64) int {
	var last20 = copyLast(series, 20)
	var medians = Ewma(last20, 10)
	var median = medians[len(medians)-1]
	var stdev = EwmStdLast(last20, 10)

	var ret = median + factor*stdev + 1000
	if math.IsNaN(ret) {
		return 0
	}
	return int(ret)
}
