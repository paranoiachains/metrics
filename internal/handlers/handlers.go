package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/paranoiachains/metrics/internal/storage"
	"github.com/paranoiachains/metrics/internal/utils"
)

// changes value of global storage, returns status code
func updateMetric(r *http.Request, metricType string) (int, error) {
	// only post methods
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, nil
	}

	metricName, metricValue, err := utils.ConvertURL(r, metricType)
	if err != nil {
		if err == utils.ErrURLFormat {
			return http.StatusBadRequest, err
		}
		return http.StatusNotFound, err
	}

	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return http.StatusBadRequest, err
		}
		storage.Storage.Gauge[metricName] = v

	case "counter":
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return http.StatusBadRequest, err
		}
		storage.Storage.Counter[metricName] += v
	}

	return http.StatusOK, nil
}

// middleware
func MetricHandler(metricType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		status, err := updateMetric(r, metricType)
		if err != nil {
			log.Println(err)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
	}
}

func LogRequest(r *http.Request) {
	log.Printf("request method: %v", r.Method)
	log.Printf("request path: %v", r.URL.Path)
	log.Printf("request content-type: %v", r.Header.Get("Content-Type"))
	log.Printf("request content-length: %v\n\n", r.Header.Get("Content-Length"))
}
