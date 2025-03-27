package handlers

import (
	"fmt"

	"github.com/paranoiachains/metrics/internal/collector"
)

// store values (temporary choice)
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

type Database interface {
	Update(mtype string, id string, value any)
	Return(mtype string, id string) (*collector.Metric, error)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (s *MemStorage) Clear() {
	s.Gauge = make(map[string]float64)
	s.Counter = make(map[string]int64)
}

func (s *MemStorage) Update(mtype string, id string, value any) {
	switch mtype {
	case "gauge":
		v, ok := value.(float64)
		if !ok {
			return
		}
		s.Gauge[id] = v

	case "counter":
		v, ok := value.(int64)
		if !ok {
			return
		}
		s.Counter[id] += v
	}
}

func (s MemStorage) Return(mtype string, id string) (*collector.Metric, error) {
	switch mtype {
	case "gauge":
		v, ok := s.Gauge[id]
		if !ok {
			return nil, fmt.Errorf("no such gauge metric")
		}
		return &collector.Metric{ID: id, MType: mtype, Value: &v}, nil

	case "counter":
		v, ok := s.Counter[id]
		if !ok {
			return nil, fmt.Errorf("no such counter metric")
		}
		return &collector.Metric{ID: id, MType: mtype, Delta: &v}, nil
	}

	return &collector.Metric{ID: id, MType: mtype}, fmt.Errorf("unknown metric type")
}

var Storage = NewMemStorage()
