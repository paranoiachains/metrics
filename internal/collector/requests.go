package collector

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
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

	// adding signature header if flag provided
	if flags.ClientKey != "" {
		h := hmac.New(sha256.New, []byte(flags.ClientKey))
		h.Write(obj)
		token := h.Sum(nil)
		hexHash := hex.EncodeToString(token)
		req.Header.Set("HashSHA256", hexHash)
	}

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response! got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	return nil
}

// Send HTTP requests with collected metrics
func Send(endpoint string) error {
	mu.Lock()
	defer mu.Unlock()

	obj, err := json.Marshal(MyMetrics)
	if err != nil {
		return err
	}

	err = NewRequest(fmt.Sprintf("http://%s/updates/", endpoint), obj)
	if err != nil {
		return err
	}
	return nil
}

// Send HTTP requests with collected metrics with interval
func SendWithInterval(reportInterval int, endpoint string) error {
	for {
		time.Sleep(time.Duration(reportInterval) * time.Second)

		var lastErr error
		retryDelays := []time.Duration{1, 3, 5}

		for _, delay := range retryDelays {
			if err := Send(endpoint); err == nil {
				fmt.Println("Metrics sent!")
				lastErr = nil
				break
			} else {
				lastErr = err
				fmt.Printf("Send failed, retrying in %v...\n", delay*time.Second)
				time.Sleep(delay * time.Second)
			}
		}

		if lastErr != nil {
			fmt.Println("error: ", lastErr)
			log.Fatal("No connection to server, exiting program")
		}
	}
}
