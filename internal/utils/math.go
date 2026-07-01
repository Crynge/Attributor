package utils

import "math"

func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func Logit(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return math.Log(p / (1 - p))
}

func Softmax(vals []float64) []float64 {
	max := vals[0]
	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	var sum float64
	out := make([]float64, len(vals))
	for i, v := range vals {
		out[i] = math.Exp(v - max)
		sum += out[i]
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}

func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

func Combinations(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	return Factorial(n) / (Factorial(k) * Factorial(n-k))
}
