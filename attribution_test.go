package main

import (
	"math"
	"testing"

	"github.com/Crynge/Attributor/internal/attribution"
)

func TestFirstTouch(t *testing.T) {
	model := &attribution.FirstTouchModel{}
	paths := [][]string{
		{"email", "search", "social"},
		{"search", "display"},
		{"social", "email", "search"},
	}
	conv := []bool{true, true, false}
	results := model.Attribute(paths, conv)
	credits := mapResults(results)
	if credits["email"] != 1 {
		t.Errorf("expected email credit 1, got %f", credits["email"])
	}
	if credits["search"] != 1 {
		t.Errorf("expected search credit 1, got %f", credits["search"])
	}
	if credits["social"] != 0 {
		t.Errorf("expected social credit 0, got %f", credits["social"])
	}
}

func TestLastTouch(t *testing.T) {
	model := &attribution.LastTouchModel{}
	paths := [][]string{
		{"email", "search", "social"},
		{"search", "display"},
		{"social", "email", "search"},
	}
	conv := []bool{true, true, false}
	results := model.Attribute(paths, conv)
	credits := mapResults(results)
	if credits["social"] != 1 {
		t.Errorf("expected social credit 1, got %f", credits["social"])
	}
	if credits["display"] != 1 {
		t.Errorf("expected display credit 1, got %f", credits["display"])
	}
}

func TestLinear(t *testing.T) {
	model := &attribution.LinearModel{}
	paths := [][]string{
		{"a", "b"},
		{"a", "b", "c"},
	}
	conv := []bool{true, true}
	results := model.Attribute(paths, conv)
	credits := mapResults(results)
	if math.Abs(credits["a"]-0.5-1.0/3.0) > 1e-6 {
		t.Errorf("unexpected a credit: %f", credits["a"])
	}
}

func TestTimeDecay(t *testing.T) {
	model := &attribution.TimeDecayModel{}
	paths := [][]string{
		{"email", "search"},
	}
	conv := []bool{true}
	results := model.Attribute(paths, conv)
	credits := mapResults(results)
	if credits["search"] <= credits["email"] {
		t.Errorf("expected last touch (search) to have higher credit than first (email): search=%f email=%f", credits["search"], credits["email"])
	}
}

func TestShapleyValue(t *testing.T) {
	model := &attribution.ShapleyValueModel{}
	paths := [][]string{
		{"a", "b"},
		{"a"},
		{"b"},
	}
	conv := []bool{true, true, false}
	results := model.Attribute(paths, conv)
	credits := mapResults(results)
	total := credits["a"] + credits["b"]
	if math.Abs(total-2) > 1e-6 {
		t.Errorf("expected total credit 2, got %f", total)
	}
}

func TestShapleySymmetric(t *testing.T) {
	model := &attribution.ShapleyValueModel{}
	paths := [][]string{
		{"a", "b"},
		{"b", "a"},
	}
	conv := []bool{true, true}
	results := model.Attribute(paths, conv)
	credits := mapResults(results)
	if math.Abs(credits["a"]-credits["b"]) > 1e-6 {
		t.Errorf("expected symmetric credits: a=%f b=%f", credits["a"], credits["b"])
	}
}

func TestMarkovChain(t *testing.T) {
	model := &attribution.MarkovChainModel{}
	paths := [][]string{
		{"a", "b"},
		{"a"},
		{"b"},
	}
	conv := []bool{true, true, false}
	results := model.Attribute(paths, conv)
	if len(results) == 0 {
		t.Fatal("expected non-empty results")
	}
	total := 0.0
	for _, r := range results {
		total += r.Share
	}
	if math.Abs(total-1) > 1e-6 {
		t.Errorf("expected shares sum to 1, got %f", total)
	}
}

func TestShapleyRemoveEffect(t *testing.T) {
	conversions := map[int]float64{
		0: 0,
		1: 5,
		2: 3,
		3: 10,
	}
	effects := attribution.RemoveEffect(conversions, 10, 2)
	if len(effects) != 2 {
		t.Fatalf("expected 2 effects, got %d", len(effects))
	}
}

func mapResults(results []attribution.AttributionResult) map[string]float64 {
	m := make(map[string]float64)
	for _, r := range results {
		m[r.Channel] = r.Credit
	}
	return m
}
