package utils

import (
	"errors"
	"log"
	"net/http"
	"strings"
)

var (
	ErrMetricType = errors.New("convert_url: metric type error")
	ErrMetricVal  = errors.New("convert_url: metric val error")
	ErrURLFormat  = errors.New("convert_url: invalid url format")
	ErrNoName     = errors.New("convert_url: no metric name")
)

// convert /update/gauge/var/123 to metricName = var; metricValue = 123
func ConvertURL(r *http.Request, metricType string) (string, string, error) {
	log.Printf("req: %v", r)
	log.Printf("path: %v", r.URL.Path)

	prefix := "/update/" + metricType + "/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		return "", "", ErrURLFormat
	}

	url := strings.TrimPrefix(r.URL.Path, prefix)
	metricName, metricValue, found := strings.Cut(url, "/")
	if !found || metricName == "" || metricValue == "" {
		return "", "", ErrNoName
	}

	return metricName, metricValue, nil
}
