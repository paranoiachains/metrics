package collector

import (
	"fmt"
	"net/http"
	"time"
)

// POST request wrapper
func NewRequest(url string) error {
	r, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("metrics_new_request: bad response! got %v, want %v", r.StatusCode, http.StatusOK)
	}
	defer r.Body.Close()
	return nil
}

// Send HTTP requests with collected metrics
func Send(endpoint string) {
	mu.Lock()
	defer mu.Unlock()
	for k, v := range MyMetrics.Gauge {
		err := NewRequest(fmt.Sprintf("http://%s/update/%s/%s/%v", endpoint, "gauge", k, v))
		if err != nil {
			fmt.Println("send_metrics: error sending gauge metric:", err)
		}
	}
	for k, v := range MyMetrics.Counter {
		err := NewRequest(fmt.Sprintf("http://%s/update/%s/%s/%v", endpoint, "counter", k, v))
		if err != nil {
			fmt.Println("send_metrics: error sending counter metric:", err)
		}
	}
}

// Send HTTP requests with collected metrics with interval
func SendWithInterval(reportInterval int, endpoint string) {
	for {
		time.Sleep(time.Duration(reportInterval) * time.Second)
		Send(endpoint)
		fmt.Println("Metrics sent!")
	}
}
