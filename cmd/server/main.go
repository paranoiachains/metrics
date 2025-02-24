package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrMetricType = errors.New("convert_url: metric type error")
	ErrMetricVal  = errors.New("convert_url: metric val error")
	ErrURLFormat  = errors.New("convert_url: invalid url format")
)

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

var storage = NewMemStorage()

func convertURL(r *http.Request, metricType string) (string, string, error) {
	log.Printf("req: %v", r)
	log.Printf("path: %v", r.URL.Path)

	prefix := "/update/" + metricType + "/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		return "", "", ErrURLFormat
	}

	url := strings.TrimPrefix(r.URL.Path, prefix)
	metricName, metricValue, found := strings.Cut(url, "/")
	if !found || metricName == "" || metricValue == "" {
		return "", "", ErrURLFormat
	}

	return metricName, metricValue, nil
}

func updateMetric(r *http.Request, metricType string) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, nil
	}

	metricName, metricValue, err := convertURL(r, metricType)
	if err != nil {
		if err == ErrMetricVal {
			return http.StatusBadRequest, err
		}
		return http.StatusNotFound, err
	}

	log.Printf("metricName: %s, metricValue: %s", metricName, metricValue)

	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return http.StatusBadRequest, err
		}
		storage.Gauge[metricName] = v

	case "counter":
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return http.StatusBadRequest, err
		}
		storage.Counter[metricName] += v
	}

	return http.StatusOK, nil
}

func metricHandler(metricType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updateMetric(r, metricType)
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(status)
	}
}

func main() {
	storage.Clear()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/", metricHandler("gauge"))
	mux.HandleFunc("/update/counter/", metricHandler("counter"))
	mux.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("bad request")
	})

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}
