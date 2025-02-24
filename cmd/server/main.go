package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var ErrMetricType = errors.New("convert_url: metric type error")
var ErrMetricVal = errors.New("convert_url: metric val error")
var ErrURLFormat = errors.New("convert_url: invalid url format")

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

var storage = NewMemStorage()

func convertURL(r *http.Request, metricType string) (metricName, metricValue string, err error) {
	log.Printf("req: 	%v", r)
	log.Printf("path: 	%v", r.URL.Path)

	var prefix string
	switch metricType {
	case "gauge":
		prefix = "/update/gauge/"
	case "counter":
		prefix = "/update/counter/"
	default:
		return "", "", ErrMetricType
	}

	url, found := strings.CutPrefix(r.URL.Path, prefix)
	if !found {
		return "", "", ErrURLFormat
	}

	metrics := strings.Split(url, "/")
	if len(metrics) < 2 {
		return "", "", ErrURLFormat
	}

	if !strings.ContainsAny(metrics[1], "0123456789") {
		return "", "", ErrMetricVal
	}

	return metrics[0], metrics[1], nil
}

func GaugeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	metricName, metricValue, err := convertURL(r, "gauge")
	if err != nil {
		if err == ErrMetricVal {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("bad request")
			return
		}
		if err == ErrURLFormat {
			w.WriteHeader(http.StatusNotFound)
			log.Println("not found")
			return
		}
	}
	log.Printf("metricName: %s, metricValue: %s", metricName, metricValue)
	v, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	storage.Gauge[metricName] = v
	w.WriteHeader(http.StatusOK)

}

func CounterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	metricName, metricValue, err := convertURL(r, "counter")
	if err != nil {
		if err == ErrMetricVal {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("bad request")
			return
		}
		if err == ErrURLFormat {
			w.WriteHeader(http.StatusNotFound)
			log.Println("not found")
			return
		}
	}

	log.Printf("metricName: %s, metricValue: %s", metricName, metricValue)
	v, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("status not found")
		return
	}
	storage.Counter[metricName] += v
	w.WriteHeader(http.StatusOK)
	log.Printf("counterValue: %d", storage.Counter[metricName])
}

func (s *MemStorage) Clear() {
	clear(s.Counter)
	clear(s.Gauge)
}

func main() {
	storage.Clear()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/", GaugeHandler)
	mux.HandleFunc("/update/counter/", CounterHandler)
	mux.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("bad request")
	})

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}
