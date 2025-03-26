package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// POST request wrapper
func NewRequest(url string, obj []byte) error {
	r, err := http.Post(url, "application/json", bytes.NewBufferString(string(obj)))
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("metrics_new_request: bad response! got %v, want %v", r.StatusCode, http.StatusOK)
	}
	r.Body.Close()

	return nil
}

// Send HTTP requests with collected metrics
func Send(endpoint string) {
	mu.Lock()
	defer mu.Unlock()
	for i := range MyMetrics {
		obj, _ := json.Marshal(MyMetrics[i])
		err := NewRequest(fmt.Sprintf("http://%s/update", endpoint), obj)
		if err != nil {
			return
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
