package sender

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type Metrics struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

var (
	MyMetrics = Metrics{Gauge: make(map[string]float64), Counter: make(map[string]int64)}
	mu        sync.Mutex
)

func (m *Metrics) Clear() {
	m.Gauge = make(map[string]float64)
	m.Counter = make(map[string]int64)
}

func GetMemStats() map[string]float64 {
	metrics := runtime.MemStats{}
	runtime.ReadMemStats(&metrics)

	m := map[string]float64{
		"Alloc":         float64(metrics.Alloc),
		"TotalAlloc":    float64(metrics.TotalAlloc),
		"Sys":           float64(metrics.Sys),
		"Lookups":       float64(metrics.Lookups),
		"Mallocs":       float64(metrics.Mallocs),
		"Frees":         float64(metrics.Frees),
		"HeapAlloc":     float64(metrics.HeapAlloc),
		"HeapSys":       float64(metrics.HeapSys),
		"HeapIdle":      float64(metrics.HeapIdle),
		"HeapInuse":     float64(metrics.HeapInuse),
		"HeapReleased":  float64(metrics.HeapReleased),
		"HeapObjects":   float64(metrics.HeapObjects),
		"StackInuse":    float64(metrics.StackInuse),
		"StackSys":      float64(metrics.StackSys),
		"MSpanInuse":    float64(metrics.MSpanInuse),
		"MSpanSys":      float64(metrics.MSpanSys),
		"MCacheInuse":   float64(metrics.MCacheInuse),
		"MCacheSys":     float64(metrics.MCacheSys),
		"BuckHashSys":   float64(metrics.BuckHashSys),
		"GCSys":         float64(metrics.GCSys),
		"OtherSys":      float64(metrics.OtherSys),
		"NextGC":        float64(metrics.NextGC),
		"LastGC":        float64(metrics.LastGC),
		"PauseTotalNs":  float64(metrics.PauseTotalNs),
		"NumGC":         float64(metrics.NumGC),
		"NumForcedGC":   float64(metrics.NumForcedGC),
		"GCCPUFraction": metrics.GCCPUFraction,
	}
	return m
}

func UpdateWithInterval(pollInterval int) {
	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)
		m := GetMemStats()
		CompareGauge(m)
		UpdateGauge(m)
		fmt.Println(MyMetrics.Gauge)
		fmt.Println(MyMetrics.Counter)
	}
}

func MetricsNewRequest(url string) error {
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

func SendMetrics() {
	mu.Lock()
	defer mu.Unlock()
	for k, v := range MyMetrics.Gauge {
		err := MetricsNewRequest(fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", "gauge", k, v))
		if err != nil {
			fmt.Println("send_metrics: error sending gauge metric:", err)
		}
	}
	for k, v := range MyMetrics.Counter {
		err := MetricsNewRequest(fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", "counter", k, v))
		if err != nil {
			fmt.Println("send_metrics: error sending counter metric:", err)
		}
	}
}

func SendMetricsWithInterval(reportInterval int) {
	for {
		time.Sleep(time.Duration(reportInterval) * time.Second)
		SendMetrics()
		fmt.Println("Sent!")
	}

}

func CompareGauge(m map[string]float64) {
	mu.Lock()
	defer mu.Unlock()
	if len(MyMetrics.Gauge) == 0 {
		return
	}
	for k, _ := range m {
		if m[k] != MyMetrics.Gauge[k] {
			MyMetrics.Counter["PollCount"] += 1
		}
	}
}

func UpdateRandomValue(v float64) {
	MyMetrics.Gauge["RandomValue"] = v
}

func UpdateGauge(m map[string]float64) {
	mu.Lock()
	defer mu.Unlock()
	for k, v := range m {
		MyMetrics.Gauge[k] = v
	}
	UpdateRandomValue(rand.Float64())
}
