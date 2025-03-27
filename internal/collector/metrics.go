package collector

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

var (
	MyMetrics Metrics
	mu        sync.Mutex
	PollCount int64 = 0
)

type Metrics []Metric

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m Metric) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("id", m.ID)
	enc.AddString("type", m.MType)
	if m.Value != nil {
		enc.AddFloat64("value", *m.Value)
	}
	if m.Delta != nil {
		enc.AddInt64("delta", *m.Delta)
	}
	return nil
}

// fetch runtime stats
func GetRuntimeStats() map[string]float64 {
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

// increase PollCount in Counter if runtime metrics have changed
func CompareGauge(m map[string]float64) {
	mu.Lock()
	defer mu.Unlock()
	if len(MyMetrics) == 0 {
		return
	}
	var delta int64 = 0
	i := 0
	for k := range m {
		if m[k] != *MyMetrics[i].Value {
			i++
			delta++
		}
	}
	PollCount += delta
}

// update Gauge
func UpdateMetrics(m map[string]float64) {
	mu.Lock()
	defer mu.Unlock()
	ClearMetrics()

	for k, v := range m {
		MyMetrics = append(MyMetrics, Metric{
			ID:    k,
			MType: "gauge",
			Value: &v,
		})
	}

	r := rand.Float64()
	MyMetrics = append(MyMetrics, Metric{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &r,
	})

	pc := PollCount
	MyMetrics = append(MyMetrics, Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: &pc,
	})
}

func ClearMetrics() {
	MyMetrics = make(Metrics, 0)
}

// update metrics storage with interval
func UpdateWithInterval(pollInterval int) {
	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)
		// 1. fetch runtime metrics
		m := GetRuntimeStats()
		// 2. check what've changed
		CompareGauge(m)
		// 3. update metrics storage
		UpdateMetrics(m)
		fmt.Println("Metrics updated.")
	}
}
