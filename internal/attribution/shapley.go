package attribution

type ShapleyValueModel struct{}

func (m *ShapleyValueModel) Name() string { return "shapley_value" }

func (m *ShapleyValueModel) Attribute(journeys [][]string, converted []bool) []AttributionResult {
	channels := collectChannels(journeys)
	n := len(channels)
	if n > 8 {
		panic("shapley: too many channels (max 8)")
	}
	if n == 0 {
		return nil
	}
	index := make(map[string]int)
	for i, ch := range channels {
		index[ch] = i
	}
	conversions := precomputeConversions(journeys, converted, index, n)
	shapley := make([]float64, n)
	for i := 0; i < n; i++ {
		total := 0.0
		for mask := 0; mask < (1 << n); mask++ {
			if mask&(1<<i) == 0 {
				continue
			}
			subset := mask & ^(1<<i)
			size := float64(popcount(subset))
			weight := factorial(size) * factorial(float64(n)-size-1) / factorial(float64(n))
			total += weight * (conversions[mask] - conversions[subset])
		}
		shapley[i] = total
	}
	totalShapley := 0.0
	for _, v := range shapley {
		totalShapley += v
	}
	var results []AttributionResult
	for i, ch := range channels {
		share := 0.0
		if totalShapley > 0 {
			share = shapley[i] / totalShapley
		}
		results = append(results, AttributionResult{
			Channel: ch,
			Credit:  shapley[i],
			Share:   share,
		})
	}
	return results
}

func precomputeConversions(journeys [][]string, converted []bool, index map[string]int, n int) map[int]float64 {
	m := 1 << n
	conv := make(map[int]float64, m)
	for mask := 0; mask < m; mask++ {
		conv[mask] = 0
	}
	for idx, j := range journeys {
		if !converted[idx] {
			continue
		}
		mask := 0
		for _, ch := range j {
			if i, ok := index[ch]; ok {
				mask |= 1 << i
			}
		}
		conv[mask]++
	}
	return conv
}

func popcount(x int) int {
	c := 0
	for x != 0 {
		x &= x - 1
		c++
	}
	return c
}

func factorial(n float64) float64 {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func Combinations(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	return int(factorial(float64(n)) / (factorial(float64(k)) * factorial(float64(n-k))))
}

func RemoveEffect(conversions map[int]float64, totalConversions float64, n int) []float64 {
	if totalConversions == 0 {
		totalConversions = 1
	}
	fullMask := (1 << n) - 1
	effects := make([]float64, n)
	for i := 0; i < n; i++ {
		withChannel := conversions[fullMask]
		withoutChannel := conversions[fullMask & ^(1<<i)]
		effects[i] = (withChannel - withoutChannel) / totalConversions
	}
	return effects
}
