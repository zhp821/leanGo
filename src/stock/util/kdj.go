package util

import (
	"stock/pojo"
)

type Kdj struct {
	n1 int
	n2 int
	n3 int
}

// NewKdj(9, 3, 3)
func NewKdj(n1 int, n2 int, n3 int) *Kdj {
	return &Kdj{n1: n1, n2: n2, n3: n3}
}

func (this *Kdj) maxHigh(bids []pojo.Stock) (h float32) {
	h = bids[0].High
	for i := 0; i < len(bids); i++ {
		if bids[i].High > h {
			h = bids[i].High
		}
	}
	return
}

func (this *Kdj) minLow(bids []pojo.Stock) (l float32) {
	l = bids[0].Low
	for i := 0; i < len(bids); i++ {
		if bids[i].Low < l {
			l = bids[i].Low
		}
	}
	return
}

func getAvg(arr []float32) (avg float32) {

	size := len(arr)
	var sum float32

	for i := 0; i < size; i++ {
		sum = sum + arr[i]
	}

	avg = sum / float32(size)

	return avg
}

func (this *Kdj) sma(x []float32, n float32) (r []float32) {
	r = make([]float32, len(x))
	for i := 0; i < len(x); i++ {
		if i < 3 {
			r[i] = 0
			continue
		}
		if i == 3 {
			r[i] = getAvg(x[0:3])
		} else {
			r[i] = (1.0*x[i] + (n-1.0)*r[i-1]) / n
		}

	}
	return
}

func (this *Kdj) Kdj(bids []pojo.Stock) (k, d, j, rsv []float32) {
	l := len(bids)
	if l < 9 {
		return
	}
	rsv = make([]float32, l)
	j = make([]float32, l)
	rsv[0] = 0
	for i := 1; i <= l; i++ {
		if i < 9 {
			rsv[i] = 0
			continue
		}
		m := i - this.n1
		if m < 0 {
			m = 0
		}
		h := this.maxHigh(bids[m:i])
		l := this.minLow(bids[m:i])

		rsv[i-1] = (bids[i-1].Close - l) * 100.0 / (h - l)
		rsv[i-1] = rsv[i-1]
	}

	k = this.sma(rsv, float32(this.n2))
	d = this.sma(k, float32(this.n3))
	for i := 0; i < l; i++ {
		j[i] = 3.0*k[i] - 2.0*d[i]
	}
	return
}
func MinAndMax(array []pojo.Stock) (float32, float32) {
	if len(array) < 1 {
		return 0, 0
	}
	min := array[0].Low
	max := array[0].High
	for _, v := range array {
		if v.Low < min {
			min = v.Low
		}
		if v.High > max {
			max = v.High
		}
	}
	return min, max
}
