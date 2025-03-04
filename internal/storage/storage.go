package storage

import "fmt"

// store values (temporary choice)
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
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

func (s MemStorage) LogStorage() {
	fmt.Printf("Storage state:\n\n, Gauge: %v\n\n, Counter: %v\n\n", s.Gauge, s.Counter)
	fmt.Println(s.Gauge)
	fmt.Println(s.Counter)
}

var Storage = NewMemStorage()
