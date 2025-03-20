package handlers

// store values (temporary choice)
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

type Database interface {
	Update(valueType string, name string, value any)
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

func (s *MemStorage) Update(valueType string, name string, value any) {
	switch valueType {
	case "gauge":
		value := value.(float64)
		s.Gauge[name] = value

	case "counter":
		value := value.(int64)
		s.Counter[name] += value
	}
}

var Storage = NewMemStorage()
