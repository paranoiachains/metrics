package collector

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

var (
	MyMetrics = Metrics{Gauge: make(map[string]float64), Counter: make(map[string]int64)}
	mu        sync.Mutex
)

// storage struct
type Metrics struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

// fetch runtime stats
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

// increase PollCount in Counter if runtime metrics have changed
func CompareGauge(m map[string]float64) {
	mu.Lock()
	defer mu.Unlock()
	if len(MyMetrics.Gauge) == 0 {
		return
	}
	for k := range m {
		if m[k] != MyMetrics.Gauge[k] {
			MyMetrics.Counter["PollCount"] += 1
		}
	}
}

// update Gauge
func UpdateGauge(m map[string]float64) {
	mu.Lock()
	defer mu.Unlock()
	for k, v := range m {
		MyMetrics.Gauge[k] = v
	}
	// random value gauge variable
	MyMetrics.Gauge["RandomValue"] = rand.Float64()
}

// update metrics storage with interval
func UpdateWithInterval(pollInterval int) {
	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)
		// 1. fetch runtime metrics
		m := GetMemStats()
		// 2. check what've changed
		CompareGauge(m)
		// 3. update metrics storage
		UpdateGauge(m)
		fmt.Println("Metrics updated.")
	}
}
