package mmm

import (
	"math"
)

type MMMParams struct {
	Alpha   []float64 `json:"alpha"`
	Beta    []float64 `json:"beta"`
	K       []float64 `json:"k"`
	Decay   []float64 `json:"decay"`
	Intercept float64 `json:"intercept"`
	Sigma   float64   `json:"sigma"`
}

type MediaMixModel struct {
	nChannels int
	Params    MMMParams
}

func NewMediaMixModel(nChannels int) *MediaMixModel {
	return &MediaMixModel{
		nChannels: nChannels,
		Params: MMMParams{
			Alpha:     make([]float64, nChannels),
			Beta:      make([]float64, nChannels),
			K:         make([]float64, nChannels),
			Decay:     make([]float64, nChannels),
			Intercept: 0,
			Sigma:     1.0,
		},
	}
}

func hill(x, alpha, beta, k float64) float64 {
	if x < 0 {
		return 0
	}
	return beta * math.Pow(x, alpha) / (math.Pow(k, alpha) + math.Pow(x, alpha))
}

func adstock(xs []float64, decay float64) []float64 {
	out := make([]float64, len(xs))
	var carry float64
	for i, x := range xs {
		out[i] = x + decay*carry
		carry = out[i]
	}
	return out
}

func (m *MediaMixModel) Predict(spends [][]float64) []float64 {
	n := len(spends)
	if n == 0 {
		return nil
	}
	pred := make([]float64, n)
	for i := 0; i < n; i++ {
		var sum float64
		for c := 0; c < m.nChannels; c++ {
			transformed := adstock([]float64{spends[i][c]}, m.Params.Decay[c])
			sum += hill(transformed[0], m.Params.Alpha[c], m.Params.Beta[c], m.Params.K[c])
		}
		pred[i] = m.Params.Intercept + sum
	}
	return pred
}

func (m *MediaMixModel) ROI(spends [][]float64) []float64 {
	if len(spends) == 0 {
		return nil
	}
	rois := make([]float64, m.nChannels)
	totalSpend := make([]float64, m.nChannels)
	for _, row := range spends {
		for c := 0; c < m.nChannels; c++ {
			totalSpend[c] += row[c]
		}
	}
	for c := 0; c < m.nChannels; c++ {
		if totalSpend[c] <= 0 {
			continue
		}
		var totalContribution float64
		for _, row := range spends {
			transformed := adstock([]float64{row[c]}, m.Params.Decay[c])
			totalContribution += hill(transformed[0], m.Params.Alpha[c], m.Params.Beta[c], m.Params.K[c])
		}
		rois[c] = totalContribution / totalSpend[c]
	}
	return rois
}
