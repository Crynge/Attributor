package attribution

import "math"

type MarkovChainModel struct{}

func (m *MarkovChainModel) Name() string { return "markov_chain" }

func (m *MarkovChainModel) Attribute(journeys [][]string, converted []bool) []AttributionResult {
	channels := collectChannels(journeys)
	n := len(channels)
	if n == 0 {
		return nil
	}
	index := make(map[string]int)
	for i, ch := range channels {
		index[ch] = i
	}
	states := n + 2
	startIdx := n
	convIdx := n + 1
	transitionCount := make([][]float64, states)
	for i := range transitionCount {
		transitionCount[i] = make([]float64, states)
	}
	for i, j := range journeys {
		if !converted[i] {
			continue
		}
		if len(j) == 0 {
			continue
		}
		prev := startIdx
		for _, ch := range j {
			curr, ok := index[ch]
			if !ok {
				continue
			}
			transitionCount[prev][curr]++
			prev = curr
		}
		transitionCount[prev][convIdx]++
	}
	transitionProb := make([][]float64, states)
	for i := range transitionProb {
		transitionProb[i] = make([]float64, states)
		if i == convIdx {
			transitionProb[i][i] = 1.0
			continue
		}
		total := 0.0
		for k := 0; k < states; k++ {
			total += transitionCount[i][k]
		}
		if total > 0 {
			for k := 0; k < states; k++ {
				transitionProb[i][k] = transitionCount[i][k] / total
			}
		}
	}
	reach := make([]float64, states)
	reach[startIdx] = 1.0
	for iter := 0; iter < 100; iter++ {
		newReach := make([]float64, states)
		for i := 0; i < states; i++ {
			if reach[i] == 0 {
				continue
			}
			for j := 0; j < states; j++ {
				newReach[j] += reach[i] * transitionProb[i][j]
			}
		}
		reach = newReach
	}
	base := reach[convIdx]
	removalEffects := make([]float64, n)
	for i := 0; i < n; i++ {
		tempProb := copyMatrix(transitionProb)
		for s := 0; s < states; s++ {
			tempProb[s][i] = 0
		}
		tempReach := make([]float64, states)
		tempReach[startIdx] = 1.0
		for iter := 0; iter < 100; iter++ {
			newReach := make([]float64, states)
			for s := 0; s < states; s++ {
				if tempReach[s] == 0 {
					continue
				}
				for t := 0; t < states; t++ {
					newReach[t] += tempReach[s] * tempProb[s][t]
				}
			}
			tempReach = newReach
		}
		removalEffects[i] = (base - tempReach[convIdx]) / math.Max(base, 1e-10)
	}
	totalEffect := 0.0
	for _, v := range removalEffects {
		totalEffect += v
	}
	var results []AttributionResult
	for i, ch := range channels {
		share := 0.0
		if totalEffect > 0 {
			share = removalEffects[i] / totalEffect
		}
		results = append(results, AttributionResult{
			Channel: ch,
			Credit:  removalEffects[i],
			Share:   share,
		})
	}
	return results
}

func copyMatrix(m [][]float64) [][]float64 {
	out := make([][]float64, len(m))
	for i := range m {
		out[i] = make([]float64, len(m[i]))
		copy(out[i], m[i])
	}
	return out
}
