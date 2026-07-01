package attribution

type AttributionResult struct {
	Channel string  `json:"channel"`
	Credit  float64 `json:"credit"`
	Share   float64 `json:"share"`
}

type AttributionModel interface {
	Name() string
	Attribute(journeys [][]string, converted []bool) []AttributionResult
}

type FirstTouchModel struct{}

func (m *FirstTouchModel) Name() string { return "first_touch" }

func (m *FirstTouchModel) Attribute(journeys [][]string, converted []bool) []AttributionResult {
	channels := collectChannels(journeys)
	credits := make(map[string]float64)
	for i, j := range journeys {
		if converted[i] && len(j) > 0 {
			credits[j[0]]++
		}
	}
	total := 0.0
	for _, v := range credits {
		total += v
	}
	return normalize(credits, channels, total)
}

type LastTouchModel struct{}

func (m *LastTouchModel) Name() string { return "last_touch" }

func (m *LastTouchModel) Attribute(journeys [][]string, converted []bool) []AttributionResult {
	channels := collectChannels(journeys)
	credits := make(map[string]float64)
	for i, j := range journeys {
		if converted[i] && len(j) > 0 {
			credits[j[len(j)-1]]++
		}
	}
	total := 0.0
	for _, v := range credits {
		total += v
	}
	return normalize(credits, channels, total)
}

type LinearModel struct{}

func (m *LinearModel) Name() string { return "linear" }

func (m *LinearModel) Attribute(journeys [][]string, converted []bool) []AttributionResult {
	channels := collectChannels(journeys)
	credits := make(map[string]float64)
	for i, j := range journeys {
		if converted[i] && len(j) > 0 {
			share := 1.0 / float64(len(j))
			for _, ch := range j {
				credits[ch] += share
			}
		}
	}
	total := 0.0
	for _, v := range credits {
		total += v
	}
	return normalize(credits, channels, total)
}

type TimeDecayModel struct{}

func (m *TimeDecayModel) Name() string { return "time_decay" }

func (m *TimeDecayModel) Attribute(journeys [][]string, converted []bool) []AttributionResult {
	channels := collectChannels(journeys)
	credits := make(map[string]float64)
	for i, j := range journeys {
		if converted[i] && len(j) > 0 {
			n := float64(len(j))
			totalWeight := 0.0
			weights := make([]float64, len(j))
			for k := range j {
				pos := float64(k)
				weights[k] = expDecay(pos, n)
				totalWeight += weights[k]
			}
			for k, ch := range j {
				credits[ch] += weights[k] / totalWeight
			}
		}
	}
	total := 0.0
	for _, v := range credits {
		total += v
	}
	return normalize(credits, channels, total)
}

func expDecay(pos, total float64) float64 {
	decay := 0.5
	return exp(decay * (pos - total + 1))
}

func exp(x float64) float64 {
	e := 1.0
	f := 1.0
	for n := 1; n < 20; n++ {
		f *= float64(n)
		e += pow(x, float64(n)) / f
	}
	return e
}

func pow(x, y float64) float64 {
	if y == 0 {
		return 1
	}
	result := 1.0
	for i := 0.0; i < y; i++ {
		result *= x
	}
	return result
}

func collectChannels(journeys [][]string) []string {
	seen := make(map[string]bool)
	var chs []string
	for _, j := range journeys {
		for _, ch := range j {
			if !seen[ch] {
				seen[ch] = true
				chs = append(chs, ch)
			}
		}
	}
	return chs
}

func normalize(credits map[string]float64, channels []string, total float64) []AttributionResult {
	if total == 0 {
		total = 1
	}
	var results []AttributionResult
	for _, ch := range channels {
		c := credits[ch]
		results = append(results, AttributionResult{
			Channel: ch,
			Credit:  c,
			Share:   c / total,
		})
	}
	return results
}
