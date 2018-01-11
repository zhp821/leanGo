package util

import (
	"strconv"
)

func ParseFloat(s string, size int) (f float32) {
	tmp, err := strconv.ParseFloat(s, 32)
	if err == nil {
		return float32(tmp)
	}
	return 0.00
}
func ParseInt(s string) (f int) {
	f, err := strconv.Atoi(s)
	if err == nil {
		return f
	}
	return 0
}
func Avg(slice []float32) float32 {
	var sum float32
	for _, v := range slice {
		sum += v
	}
	//fmt.Printf("%v and sum is %f and avg is %f", slice, sum, sum/float32(len(slice)))
	return sum / float32(len(slice))
}
func Avgint(slice []int) float32 {
	var sum int
	for _, v := range slice {
		sum += v
	}
	return (float32(sum) / float32(len(slice)))
}
