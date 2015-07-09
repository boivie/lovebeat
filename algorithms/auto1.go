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

func copyLast(series []int64, N int) []float64 {
	var start = len(series) - N
	if start < 0 {
		start = 0
	}
	var end = len(series)
	ret := make([]float64, end-start)
	for i := start; i < end; i++ {
		ret[i-start] = float64(series[i])
	}
	return ret
}

func sum(numbers []float64) (total float64) {
	for _, x := range numbers {
		total += x
	}
	return total
}

func mean(series []float64) float64 {
	return float64(sum(series)) / float64(len(series))
}

func stdDev(numbers []float64, mean float64) float64 {
	total := 0.0
	for _, number := range numbers {
		total += math.Pow(float64(number)-mean, 2)
	}
	variance := total / float64(len(numbers)-1)
	return math.Sqrt(variance)
}

func removeOutliers(series []float64, m float64) []float64 {
	u := mean(series)
	s := stdDev(series, u)
	var ret []float64

	for _, e := range series {
		if u-m*s < e && e < u+m*s {
			ret = append(ret, e)
		}
	}

	if len(ret) == 0 {
		return series
	}
	return ret
}

func autoAlgEwmaStdRemoveOutliers(series []int64, factor float64) int64 {
	var last20 = copyLast(series, 20)

	last20 = removeOutliers(last20, 3)

	var medians = Ewma(last20, 10)
	var median = medians[len(medians)-1]
	var stdev = EwmStdLast(last20, 10)

	var ret = median + factor*stdev + 1000
	if math.IsNaN(ret) {
		return 0
	}
	return int64(ret + 0.5)
}

func AutoAlg(series []int64) int64 {
	return autoAlgEwmaStdRemoveOutliers(series, 3)
}
