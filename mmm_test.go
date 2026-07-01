package main

import (
	"math"
	"testing"

	"github.com/Crynge/Attributor/internal/mmm"
)

func TestMMMPredict(t *testing.T) {
	nChannels := 2
	model := mmm.NewMediaMixModel(nChannels)
	model.Params.Alpha = []float64{1.5, 2.0}
	model.Params.Beta = []float64{2.0, 1.5}
	model.Params.K = []float64{100, 200}
	model.Params.Decay = []float64{0.3, 0.5}
	model.Params.Intercept = 10
	model.Params.Sigma = 1.0

	spends := [][]float64{
		{50, 100},
		{200, 50},
	}
	pred := model.Predict(spends)
	if len(pred) != 2 {
		t.Fatalf("expected 2 predictions, got %d", len(pred))
	}
	for i, p := range pred {
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Errorf("prediction %d is invalid: %f", i, p)
		}
	}
}

func TestMMMROI(t *testing.T) {
	nChannels := 2
	model := mmm.NewMediaMixModel(nChannels)
	model.Params.Alpha = []float64{1.0, 1.0}
	model.Params.Beta = []float64{2.0, 3.0}
	model.Params.K = []float64{100, 100}
	model.Params.Decay = []float64{0.0, 0.0}
	model.Params.Intercept = 0
	model.Params.Sigma = 1.0

	spends := [][]float64{
		{100, 100},
		{100, 100},
	}
	roi := model.ROI(spends)
	if len(roi) != 2 {
		t.Fatalf("expected 2 ROI values, got %d", len(roi))
	}
	if roi[0] <= 0 || roi[1] <= 0 {
		t.Errorf("expected positive ROI, got %f, %f", roi[0], roi[1])
	}
}

func TestMMCFit(t *testing.T) {
	nChannels := 1
	model := mmm.NewMediaMixModel(nChannels)
	spends := make([][]float64, 10)
	target := make([]float64, 10)
	for i := 0; i < 10; i++ {
		spends[i] = []float64{float64((i + 1) * 100)}
		target[i] = 500 + float64(i)*20
	}
	result := model.Fit(spends, target, 200, 50, 5, 0.1)
	if len(result.Samples) == 0 {
		t.Fatal("expected at least 1 MCMC sample")
	}
	if math.IsNaN(result.Mean.Intercept) {
		t.Error("intercept is NaN")
	}
}
