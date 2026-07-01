[![CI](https://github.com/Crynge/Attributor/actions/workflows/ci.yml/badge.svg)](https://github.com/Crynge/Attributor/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.22-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

# Attributor

**Multi-touch attribution & media mix modeling engine.**

Quantify the contribution of each marketing channel to conversions, simulate budget reallocations, and optimize spend using Shapley values, Markov chains, and Bayesian MCMC.

---

## Attribution Models

### Shapley Value Attribution

$$ \phi_i(v) = \sum_{S \subseteq N \setminus \{i\}} \frac{|S|! (|N| - |S| - 1)!}{|N|!} \left( v(S \cup \{i\}) - v(S) \right) $$

Measures the marginal contribution of each channel across all possible channel combinations — the only attribution method that satisfies symmetry, linearity, and efficiency.

### Markov Chain Attribution

```
                     ┌──────────┐
                     │  START   │
                     └────┬─────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
         ┌────▼───┐ ┌────▼───┐ ┌────▼───┐
         │ Google │ │  Meta  │ │ TikTok │
         └────┬───┘ └────┬───┘ └────┬───┘
              │           │           │
              └───────────┼───────────┘
                          │
                     ┌────▼─────┐
                     │  CONVERT │
                     └──────────┘
```

Removal effect: remove a channel and measure the drop in conversion probability. The channel with the largest drop gets the most credit.

### Bayesian MMM

$$ y_t = \beta_0 + \sum_{ch} \beta_{ch} \cdot f(x_{ch,t}) + \gamma \cdot z_t + \varepsilon_t $$

Media mix model using Hamiltonian Monte Carlo (via `golang.org/x/perf` MCMC) with:
- **Adstock transformations** — carrying over effects across weeks
- **Saturation curves** — diminishing returns via Hill functions
- **Uncertainty intervals** — 90% credible intervals for every parameter

## CLI Usage

```bash
# Run Shapley attribution
attributor attrib shapley --journeys journeys.json

# Run Markov chain attribution
attributor attrib markov --journeys journeys.json

# Fit MMM model
attributor mmm fit --data spend_revenue.csv --output model.json

# Simulate budget changes
attributor simulate --budget 500000 --reallocate

# Export report
attributor report --format json --output results.json
```

## Library

```go
import "github.com/Crynge/Attributor/internal/attribution"

shapley := attribution.NewShapleyCalculator()
result := shapley.Compute(journeys)
for ch, val := range result.Attributions {
    fmt.Printf("%s: %.2f%%\n", ch, val*100)
}

markov := attribution.NewMarkovChainCalculator()
result := markov.Compute(journeys)
for ch, val := range result.Attributions {
    fmt.Printf("%s: %.2f%%\n", ch, val*100)
}
```

## API Server

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/attribution/shapley` | Compute Shapley values |
| `POST` | `/api/v1/attribution/markov` | Compute Markov chain attribution |
| `POST` | `/api/v1/mmm/fit` | Fit MMM model |
| `POST` | `/api/v1/simulate` | Budget simulation |

## Test Coverage

```
pkg/attribution/shapley.go     95.2%
pkg/attribution/markov.go      90.8%
pkg/mmm/mcmc.go                87.3%
pkg/simulation/sim.go          92.1%
pkg/reporting/report.go        88.6%
```

## References

- Shapley, L.S. (1953). *A Value for n-Person Games.*
- Anderl, E. et al. (2016). *Mapping the Customer Journey: A Graph-Based Framework for Online Attribution.*
- Jin, Y. et al. (2017). *Bayesian Methods for Media Mix Modeling with Carryover and Shape Effects.*
