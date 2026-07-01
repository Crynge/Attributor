package mmm

import (
	"math"
	"math/rand"
)

type MCMCResult struct {
	Samples  []MMMParams
	Mean     MMMParams
	StdDev   MMMParams
	Quantiles map[float64]MMMParams
}

func (m *MediaMixModel) Fit(spends [][]float64, target []float64, iterations, burnin, thin int, scale float64) MCMCResult {
	rng := rand.New(rand.NewSource(42))
	paramCount := 4*m.nChannels + 2
	current := make([]float64, paramCount)
	for c := 0; c < m.nChannels; c++ {
		current[c] = m.Params.Alpha[c]
		current[m.nChannels+c] = m.Params.Beta[c]
		current[2*m.nChannels+c] = m.Params.K[c]
		current[3*m.nChannels+c] = m.Params.Decay[c]
	}
	current[4*m.nChannels] = m.Params.Intercept
	current[4*m.nChannels+1] = m.Params.Sigma

	ll := m.logLikelihood(current, spends, target)

	var samples []MMMParams
	for iter := 0; iter < iterations; iter++ {
		proposal := make([]float64, paramCount)
		copy(proposal, current)
		for i := range proposal {
			proposal[i] += rng.NormFloat64() * scale
		}
		propLL := m.logLikelihood(proposal, spends, target)
		logRatio := propLL - ll + m.logPrior(proposal) - m.logPrior(current)
		if logRatio > 0 || rng.Float64() < math.Exp(logRatio) {
			current = proposal
			ll = propLL
		}
		if iter >= burnin && (iter-burnin)%thin == 0 {
			samples = append(samples, m.paramsFromSlice(current))
		}
	}
	return summarize(samples)
}

func (m *MediaMixModel) logLikelihood(params []float64, spends [][]float64, target []float64) float64 {
	m.setParams(params)
	pred := m.Predict(spends)
	var ll float64
	sigma := params[4*m.nChannels+1]
	if sigma <= 0 {
		sigma = 1
	}
	for i := range target {
		diff := target[i] - pred[i]
		ll += -0.5*math.Log(2*math.Pi) - math.Log(sigma) - 0.5*diff*diff/(sigma*sigma)
	}
	return ll
}

func (m *MediaMixModel) logPrior(params []float64) float64 {
	var lp float64
	for i := 0; i < m.nChannels; i++ {
		if params[i] <= 0 {
			return -1e100
		}
		lp += -0.5 * math.Log(2*math.Pi) - math.Log(2) - 0.5*params[i]*params[i]/(4)
	}
	for i := m.nChannels; i < 2*m.nChannels; i++ {
		if params[i] <= 0 {
			return -1e100
		}
		lp += -0.5 * math.Log(2*math.Pi) - math.Log(params[i]) - 0.5*math.Log(params[i])*math.Log(params[i])
	}
	for i := 2 * m.nChannels; i < 3*m.nChannels; i++ {
		if params[i] <= 0 {
			return -1e100
		}
		lp += -0.5 * math.Log(2*math.Pi) - math.Log(params[i]) - 0.5*math.Log(params[i])*math.Log(params[i])
	}
	for i := 3 * m.nChannels; i < 4*m.nChannels; i++ {
		if params[i] <= 0 || params[i] >= 1 {
			return -1e100
		}
		lp += math.Log(1)
	}
	lp += -0.5 * math.Log(2*math.Pi) - math.Log(2) - 0.5 * params[4*m.nChannels] * params[4*m.nChannels] / 4
	if params[4*m.nChannels+1] <= 0 {
		return -1e100
	}
	lp += -0.5*math.Log(2*math.Pi) - math.Log(2) - 0.5*params[4*m.nChannels+1]*params[4*m.nChannels+1]/4
	return lp
}

func (m *MediaMixModel) setParams(params []float64) {
	for c := 0; c < m.nChannels; c++ {
		m.Params.Alpha[c] = params[c]
		m.Params.Beta[c] = params[m.nChannels+c]
		m.Params.K[c] = params[2*m.nChannels+c]
		m.Params.Decay[c] = params[3*m.nChannels+c]
	}
	m.Params.Intercept = params[4*m.nChannels]
	m.Params.Sigma = params[4*m.nChannels+1]
}

func (m *MediaMixModel) paramsFromSlice(params []float64) MMMParams {
	p := MMMParams{
		Alpha:     make([]float64, m.nChannels),
		Beta:      make([]float64, m.nChannels),
		K:         make([]float64, m.nChannels),
		Decay:     make([]float64, m.nChannels),
		Intercept: params[4*m.nChannels],
		Sigma:     params[4*m.nChannels+1],
	}
	for c := 0; c < m.nChannels; c++ {
		p.Alpha[c] = params[c]
		p.Beta[c] = params[m.nChannels+c]
		p.K[c] = params[2*m.nChannels+c]
		p.Decay[c] = params[3*m.nChannels+c]
	}
	return p
}

func summarize(samples []MMMParams) MCMCResult {
	if len(samples) == 0 {
		return MCMCResult{}
	}
	n := len(samples)
	mean := samples[0]
	for i := 1; i < n; i++ {
		mean = addParams(mean, samples[i])
	}
	mean = scaleParams(mean, 1.0/float64(n))

	var stdDev MMMParams
	nc := len(mean.Alpha)
	stdDev.Alpha = make([]float64, nc)
	stdDev.Beta = make([]float64, nc)
	stdDev.K = make([]float64, nc)
	stdDev.Decay = make([]float64, nc)
	for c := 0; c < nc; c++ {
		var sa, sb, sk, sd float64
		for i := 0; i < n; i++ {
			sa += (samples[i].Alpha[c] - mean.Alpha[c]) * (samples[i].Alpha[c] - mean.Alpha[c])
			sb += (samples[i].Beta[c] - mean.Beta[c]) * (samples[i].Beta[c] - mean.Beta[c])
			sk += (samples[i].K[c] - mean.K[c]) * (samples[i].K[c] - mean.K[c])
			sd += (samples[i].Decay[c] - mean.Decay[c]) * (samples[i].Decay[c] - mean.Decay[c])
		}
		stdDev.Alpha[c] = math.Sqrt(sa / float64(n))
		stdDev.Beta[c] = math.Sqrt(sb / float64(n))
		stdDev.K[c] = math.Sqrt(sk / float64(n))
		stdDev.Decay[c] = math.Sqrt(sd / float64(n))
	}
	var si, ss float64
	for i := 0; i < n; i++ {
		si += (samples[i].Intercept - mean.Intercept) * (samples[i].Intercept - mean.Intercept)
		ss += (samples[i].Sigma - mean.Sigma) * (samples[i].Sigma - mean.Sigma)
	}
	stdDev.Intercept = math.Sqrt(si / float64(n))
	stdDev.Sigma = math.Sqrt(ss / float64(n))

	q := map[float64]MMMParams{0.025: {}, 0.5: {}, 0.975: {}}
	for pct := range q {
		q[pct] = quantileParams(samples, pct)
	}

	return MCMCResult{
		Samples:    samples,
		Mean:       mean,
		StdDev:     stdDev,
		Quantiles:  q,
	}
}

func addParams(a, b MMMParams) MMMParams {
	r := MMMParams{
		Alpha:     make([]float64, len(a.Alpha)),
		Beta:      make([]float64, len(a.Beta)),
		K:         make([]float64, len(a.K)),
		Decay:     make([]float64, len(a.Decay)),
		Intercept: a.Intercept + b.Intercept,
		Sigma:     a.Sigma + b.Sigma,
	}
	for i := range a.Alpha {
		r.Alpha[i] = a.Alpha[i] + b.Alpha[i]
		r.Beta[i] = a.Beta[i] + b.Beta[i]
		r.K[i] = a.K[i] + b.K[i]
		r.Decay[i] = a.Decay[i] + b.Decay[i]
	}
	return r
}

func scaleParams(p MMMParams, s float64) MMMParams {
	r := MMMParams{
		Alpha:     make([]float64, len(p.Alpha)),
		Beta:      make([]float64, len(p.Beta)),
		K:         make([]float64, len(p.K)),
		Decay:     make([]float64, len(p.Decay)),
		Intercept: p.Intercept * s,
		Sigma:     p.Sigma * s,
	}
	for i := range p.Alpha {
		r.Alpha[i] = p.Alpha[i] * s
		r.Beta[i] = p.Beta[i] * s
		r.K[i] = p.K[i] * s
		r.Decay[i] = p.Decay[i] * s
	}
	return r
}

func quantileParams(samples []MMMParams, pct float64) MMMParams {
	n := len(samples)
	idx := int(pct * float64(n))
	if idx >= n {
		idx = n - 1
	}
	q := samples[idx]
	nc := len(q.Alpha)
	q.Alpha = make([]float64, nc)
	q.Beta = make([]float64, nc)
	q.K = make([]float64, nc)
	q.Decay = make([]float64, nc)
	for c := 0; c < nc; c++ {
		vals := make([]float64, n)
		for i := 0; i < n; i++ {
			vals[i] = samples[i].Alpha[c]
		}
		q.Alpha[c] = sortedQuantile(vals, pct)
		vals2 := make([]float64, n)
		for i := 0; i < n; i++ {
			vals2[i] = samples[i].Beta[c]
		}
		q.Beta[c] = sortedQuantile(vals2, pct)
		vals3 := make([]float64, n)
		for i := 0; i < n; i++ {
			vals3[i] = samples[i].K[c]
		}
		q.K[c] = sortedQuantile(vals3, pct)
		vals4 := make([]float64, n)
		for i := 0; i < n; i++ {
			vals4[i] = samples[i].Decay[c]
		}
		q.Decay[c] = sortedQuantile(vals4, pct)
	}
	is := make([]float64, n)
	ss := make([]float64, n)
	for i := 0; i < n; i++ {
		is[i] = samples[i].Intercept
		ss[i] = samples[i].Sigma
	}
	q.Intercept = sortedQuantile(is, pct)
	q.Sigma = sortedQuantile(ss, pct)
	return q
}

func sortedQuantile(vals []float64, pct float64) float64 {
	n := len(vals)
	sorted := make([]float64, n)
	copy(sorted, vals)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	idx := int(pct * float64(n-1))
	return sorted[idx]
}
