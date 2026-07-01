package journey

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Touchpoint struct {
	Channel         string    `json:"channel"`
	Timestamp       time.Time `json:"timestamp"`
	Cost            float64   `json:"cost"`
	ConversionValue float64   `json:"conversion_value"`
}

type Journey struct {
	CustomerID  string       `json:"customer_id"`
	Touchpoints []Touchpoint `json:"touchpoints"`
	Converted   bool         `json:"converted"`
}

type JourneyStore interface {
	Save(journey Journey) error
	FindByCustomerID(id string) (*Journey, error)
	All() ([]Journey, error)
}

type MemJourneyStore struct {
	journeys map[string]Journey
}

func NewMemJourneyStore() *MemJourneyStore {
	return &MemJourneyStore{journeys: make(map[string]Journey)}
}

func (s *MemJourneyStore) Save(j Journey) error {
	s.journeys[j.CustomerID] = j
	return nil
}

func (s *MemJourneyStore) FindByCustomerID(id string) (*Journey, error) {
	j, ok := s.journeys[id]
	if !ok {
		return nil, fmt.Errorf("journey not found: %s", id)
	}
	return &j, nil
}

func (s *MemJourneyStore) All() ([]Journey, error) {
	var all []Journey
	for _, j := range s.journeys {
		all = append(all, j)
	}
	return all, nil
}

type CSVJourneyStore struct {
	filePath string
	journeys map[string]Journey
}

func NewCSVJourneyStore(filePath string) (*CSVJourneyStore, error) {
	s := &CSVJourneyStore{
		filePath: filePath,
		journeys: make(map[string]Journey),
	}
	if _, err := os.Stat(filePath); err == nil {
		if err := s.load(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *CSVJourneyStore) Save(j Journey) error {
	s.journeys[j.CustomerID] = j
	return s.flush()
}

func (s *CSVJourneyStore) FindByCustomerID(id string) (*Journey, error) {
	j, ok := s.journeys[id]
	if !ok {
		return nil, fmt.Errorf("journey not found: %s", id)
	}
	return &j, nil
}

func (s *CSVJourneyStore) All() ([]Journey, error) {
	var all []Journey
	for _, j := range s.journeys {
		all = append(all, j)
	}
	return all, nil
}

func (s *CSVJourneyStore) load() error {
	f, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}
	for _, rec := range records {
		if len(rec) < 5 {
			continue
		}
		cost, _ := strconv.ParseFloat(rec[2], 64)
		convVal, _ := strconv.ParseFloat(rec[3], 64)
		converted, _ := strconv.ParseBool(rec[4])
		tp := Touchpoint{
			Channel:         rec[0],
			Cost:            cost,
			ConversionValue: convVal,
		}
		customerID := rec[0]
		if existing, ok := s.journeys[customerID]; ok {
			existing.Touchpoints = append(existing.Touchpoints, tp)
			s.journeys[customerID] = existing
		} else {
			s.journeys[customerID] = Journey{
				CustomerID:  customerID,
				Touchpoints: []Touchpoint{tp},
				Converted:   converted,
			}
		}
	}
	return nil
}

func (s *CSVJourneyStore) flush() error {
	f, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	for _, j := range s.journeys {
		for _, tp := range j.Touchpoints {
			w.Write([]string{
				j.CustomerID,
				tp.Channel,
				fmt.Sprintf("%f", tp.Cost),
				fmt.Sprintf("%f", tp.ConversionValue),
				fmt.Sprintf("%t", j.Converted),
			})
		}
	}
	w.Flush()
	return w.Error()
}

func (j Journey) Channels() []string {
	var chs []string
	for _, tp := range j.Touchpoints {
		chs = append(chs, tp.Channel)
	}
	return chs
}

func GenerateSampleJourneys(n int, channels []string) []Journey {
	rng := rand.New(rand.NewSource(99))
	var journeys []Journey
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("cust_%d", i)
		nTouches := rng.Intn(5) + 1
		var tps []Touchpoint
		for j := 0; j < nTouches; j++ {
			ch := channels[rng.Intn(len(channels))]
			tps = append(tps, Touchpoint{
				Channel:         ch,
				Timestamp:       time.Now().Add(-time.Duration(nTouches-j) * time.Hour),
				Cost:            rng.Float64() * 100,
				ConversionValue: rng.Float64() * 500,
			})
		}
		converted := rng.Float64() > 0.3
		journeys = append(journeys, Journey{
			CustomerID:  id,
			Touchpoints: tps,
			Converted:   converted,
		})
	}
	return journeys
}

func JourneysToTouchpointStrings(journeys []Journey) ([][]string, []bool) {
	var paths [][]string
	var conv []bool
	for _, j := range journeys {
		paths = append(paths, strings.Fields(strings.Join(j.Channels(), " ")))
		conv = append(conv, j.Converted)
	}
	return paths, conv
}
