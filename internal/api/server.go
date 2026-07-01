package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Crynge/Attributor/internal/attribution"
	"github.com/Crynge/Attributor/internal/journey"
	"github.com/Crynge/Attributor/internal/mmm"
	"github.com/Crynge/Attributor/internal/reporting"
	"github.com/Crynge/Attributor/internal/simulation"
)

type Server struct {
	journeyStore journey.JourneyStore
	mmmModel     *mmm.MediaMixModel
	simulator    *simulation.BudgetSimulator
}

func NewServer(store journey.JourneyStore) *Server {
	return &Server{
		journeyStore: store,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /attribution", s.handleAttribution)
	mux.HandleFunc("POST /mmm", s.handleMMM)
	mux.HandleFunc("POST /simulate", s.handleSimulate)
	mux.HandleFunc("GET /health", s.handleHealth)
	return mux
}

type attributionRequest struct {
	Model   string     `json:"model"`
	Journeys []journey.Journey `json:"journeys"`
}

func (s *Server) handleAttribution(w http.ResponseWriter, r *http.Request) {
	var req attributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var model attribution.AttributionModel
	switch req.Model {
	case "first_touch":
		model = &attribution.FirstTouchModel{}
	case "last_touch":
		model = &attribution.LastTouchModel{}
	case "linear":
		model = &attribution.LinearModel{}
	case "time_decay":
		model = &attribution.TimeDecayModel{}
	case "shapley_value":
		model = &attribution.ShapleyValueModel{}
	case "markov_chain":
		model = &attribution.MarkovChainModel{}
	default:
		http.Error(w, "unknown model: "+req.Model, http.StatusBadRequest)
		return
	}
	paths, conv := journey.JourneysToTouchpointStrings(req.Journeys)
	results := model.Attribute(paths, conv)
	report := reporting.AttributionReport{
		Model:   model.Name(),
		Results: results,
	}
	if b, err := json.Marshal(report); err == nil {
		reporting.ExportJSON(report, "")
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

type mmmRequest struct {
	Spends   [][]float64 `json:"spends"`
	Target   []float64   `json:"target"`
	Channels int         `json:"channels"`
	Iterations int      `json:"iterations"`
	Burnin  int          `json:"burnin"`
}

func (s *Server) handleMMM(w http.ResponseWriter, r *http.Request) {
	var req mmmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Iterations == 0 {
		req.Iterations = 10000
	}
	if req.Burnin == 0 {
		req.Burnin = 1000
	}
	model := mmm.NewMediaMixModel(req.Channels)
	result := model.Fit(req.Spends, req.Target, req.Iterations, req.Burnin, 10, 0.1)
	model.Params = result.Mean
	report := reporting.MMMReport{
		ROI:   model.ROI(req.Spends),
		Params: result.Mean,
	}
	if b, err := json.Marshal(report); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

type simulateRequest struct {
	Budget float64              `json:"budget"`
	ROI    map[string]float64   `json:"roi"`
}

func (s *Server) handleSimulate(w http.ResponseWriter, r *http.Request) {
	var req simulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sim := simulation.NewBudgetSimulator(req.Budget)
	for ch, roi := range req.ROI {
		sim.SetROI(ch, roi)
	}
	alloc := sim.Allocate()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alloc)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"version": strconv.Itoa(1),
	})
}
