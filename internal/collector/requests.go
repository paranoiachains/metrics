package collector

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/paranoiachains/metrics/internal/flags"
)

// POST request wrapper
func NewRequest(url string, obj []byte) error {
	var reqBody *bytes.Buffer
	reqBody = bytes.NewBuffer(obj)

	// if gzip encoding enabled - encode.
	if flags.EncodingEnabled {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, err := gz.Write(obj)
		if err != nil {
			return fmt.Errorf("gzip compression error: %v", err)
		}
		gz.Close()
		reqBody = &buf
	}

	// new request
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return fmt.Errorf("request creation failed: %v", err)
	}

	// setting headers
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	if flags.EncodingEnabled {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response! got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	return nil
}

// Send HTTP requests with collected metrics
func Send(endpoint string) {
	mu.Lock()
	defer mu.Unlock()

	for i := range MyMetrics {
		obj, err := json.Marshal(MyMetrics[i])
		if err != nil {
			fmt.Println("JSON marshaling error:", err)
			continue
		}

		err = NewRequest(fmt.Sprintf("http://%s/update/", endpoint), obj)
		if err != nil {
			fmt.Println("Failed to send request:", err)
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
