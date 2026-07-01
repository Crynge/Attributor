package simulation

import (
	"math"
	"sort"
)

type BudgetSimulator struct {
	ChannelROI  map[string]float64
	TotalBudget float64
	Constraints map[string]struct{ Min, Max float64 }
}

func NewBudgetSimulator(totalBudget float64) *BudgetSimulator {
	return &BudgetSimulator{
		ChannelROI:  make(map[string]float64),
		TotalBudget: totalBudget,
		Constraints: make(map[string]struct{ Min, Max float64 }),
	}
}

func (s *BudgetSimulator) SetROI(channel string, roi float64) {
	s.ChannelROI[channel] = roi
}

func (s *BudgetSimulator) SetConstraint(channel string, min, max float64) {
	s.Constraints[channel] = struct{ Min, Max float64 }{min, max}
}

func (s *BudgetSimulator) Allocate() map[string]float64 {
	type chRoi struct {
		name string
		roi  float64
	}
	var chs []chRoi
	for name, roi := range s.ChannelROI {
		chs = append(chs, chRoi{name, roi})
	}
	sort.Slice(chs, func(i, j int) bool {
		return chs[i].roi > chs[j].roi
	})

	result := make(map[string]float64)
	remaining := s.TotalBudget

	for _, ch := range chs {
		cons, hasCon := s.Constraints[ch.name]
		if hasCon {
			result[ch.name] = cons.Min
			remaining -= cons.Min
		} else {
			result[ch.name] = 0
		}
	}

	if remaining <= 0 {
		return result
	}

	for {
		allocated := false
		for _, ch := range chs {
			if remaining <= 0 {
				break
			}
			cons, hasCon := s.Constraints[ch.name]
			if hasCon && result[ch.name] >= cons.Max {
				continue
			}
			extra := math.Min(remaining, s.TotalBudget*0.1)
			if hasCon {
				extra = math.Min(extra, cons.Max-result[ch.name])
			}
			if extra > 0 {
				result[ch.name] += extra
				remaining -= extra
				allocated = true
			}
		}
		if !allocated || remaining <= 0 {
			break
		}
	}

	if remaining > 0 {
		totalROI := 0.0
		for _, ch := range chs {
			totalROI += ch.roi
		}
		if totalROI > 0 {
			for _, ch := range chs {
				share := ch.roi / totalROI
				extra := remaining * share
				cons, hasCon := s.Constraints[ch.name]
				if hasCon {
					if result[ch.name]+extra > cons.Max {
						extra = cons.Max - result[ch.name]
					}
				}
				if extra > 0 {
					result[ch.name] += extra
				}
			}
		}
	}

	return result
}

func (s *BudgetSimulator) WhatIf(channel string, budgetChange float64) map[string]float64 {
	orig := s.ChannelROI[channel]
	newROI := orig * (1 + budgetChange/100)
	s.SetROI(channel, newROI)
	return s.Allocate()
}

func (s *BudgetSimulator) Optimize(objective func(map[string]float64) float64) map[string]float64 {
	best := s.Allocate()
	bestVal := objective(best)
	for step := 0; step < 1000; step++ {
		candidate := make(map[string]float64)
		for ch := range s.ChannelROI {
			origROI := s.ChannelROI[ch]
			s.SetROI(ch, origROI*(1+(math.Sin(float64(step))*0.01)))
		}
		alloc := s.Allocate()
		val := objective(alloc)
		if val > bestVal {
			best = alloc
			bestVal = val
		}
		for ch := range s.ChannelROI {
			s.SetROI(ch, s.ChannelROI[ch])
		}
	}
	return best
}
