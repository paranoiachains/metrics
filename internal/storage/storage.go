package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/paranoiachains/metrics/internal/collector"
)

// store values (temporary choice)
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

// flexibility
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

	return nil, fmt.Errorf("unknown metric type")
}

type FileHandler interface {
	Write(filename string) error
	Read(filename string) (*collector.Metric, error)
	Restore(filename string) error
	ClearFile(filename string) error
}

func (s MemStorage) Write(filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var metrics collector.Metrics
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	for name := range s.Gauge {
		metric, err := s.Return("gauge", name)
		if err != nil {
			return err
		}
		metrics = append(metrics, *metric)
	}

	for name := range s.Counter {
		metric, err := s.Return("counter", name)
		if err != nil {
			return err
		}
		metrics = append(metrics, *metric)
	}
	if err := encoder.Encode(metrics); err != nil {
		return err
	}
	return nil
}

func (s MemStorage) Read(filename string) (*collector.Metric, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var metric collector.Metric
	if err := json.Unmarshal(data, &metric); err != nil {
		return nil, err
	}

	return &metric, nil
}

func (s *MemStorage) Restore(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var metrics collector.Metrics
		if err := decoder.Decode(&metrics); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		for _, metric := range metrics {
			switch metric.MType {
			case "gauge":
				s.Gauge[metric.ID] = *metric.Value
			case "counter":
				s.Counter[metric.ID] = *metric.Delta
			}
		}
	}
	return nil
}

func (s *MemStorage) ClearFile(filename string) error {
	if err := os.Truncate(filename, 0); err != nil {
		return err
	}
	return nil
}

func WriteWithInterval(file FileHandler, filename string, storeInterval int) {
	// lol
	if storeInterval == 0 {
		storeInterval = 1
	}
	// doesnt work without this line idk why
	time.Sleep(time.Second)
	for {
		file.ClearFile(filename)
		if err := file.Write(filename); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Duration(storeInterval) * time.Second)
	}
}

func ConnectAndPing(driverName string, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	// ping
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

var Storage = NewMemStorage()
