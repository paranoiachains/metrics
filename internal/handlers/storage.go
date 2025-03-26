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
		value := value.(float64)
		s.Gauge[id] = value

	case "counter":
		value := value.(int64)
		s.Counter[id] += value
	}
}

func (s MemStorage) Return(mtype string, id string) (*collector.Metric, error) {
	metric := new(collector.Metric)
	switch mtype {
	case "gauge":
		v, ok := s.Gauge[id]
		if !ok {
			return nil, fmt.Errorf("no such gauge metric")
		}
		metric.ID = id
		metric.MType = mtype
		metric.Value = &v

	case "counter":
		v, ok := s.Counter[id]
		if !ok {
			return nil, fmt.Errorf("no such gauge metric")
		}
		metric.ID = id
		metric.MType = mtype
		metric.Delta = &v
	}
	return metric, nil
}

var Storage = NewMemStorage()
