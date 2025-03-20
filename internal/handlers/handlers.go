package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/utils"
)

// updateMetric changes the value of global storage and returns a status code
func updateMetric(c *gin.Context, metricType string) {
	metricValue, metricName, err := utils.ExtractMetricParams(c)
	if err != nil {
		c.Status(http.StatusNotFound)
		c.Header("Content-Type", "text/plain")
		c.Header("Error", err.Error())
		return
	}

	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.Status(http.StatusBadRequest)
			c.Header("Content-Type", "text/plain")
			c.Header("Error", err.Error())
			return
		}
		Storage.Gauge[metricName] = v

	case "counter":
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.Status(http.StatusBadRequest)
			c.Header("Content-Type", "text/plain")
			c.Header("Error", err.Error())
			return
		}
		Storage.Counter[metricName] += v
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/plain")
}

// MetricHandler is a Gin route handler for POST HTTP metric updates
func MetricHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Status(http.StatusMethodNotAllowed)
			c.Header("Content-Type", "text/plain")
			return
		}
		if c.Param("metricType") != "gauge" && c.Param("metricType") != "counter" {
			c.Status(http.StatusBadRequest)
			c.Header("Content-Type", "text/plain")
			return
		}
		updateMetric(c, c.Param("metricType"))
		c.Header("Content-Type", "text/plain")
	}
}

func ReturnMetric(c *gin.Context) {
	fmt.Println("return metric init")
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")

	switch metricType {
	case "gauge":
		retrievedName, ok := Storage.Gauge[metricName]
		if !ok {
			c.Status(http.StatusNotFound)
			c.Header("Content-Type", "text/plain")
			return
		}
		c.String(200, fmt.Sprintf("%f", retrievedName))

	case "counter":
		retrievedName, ok := Storage.Counter[metricName]
		if !ok {
			c.Status(http.StatusNotFound)
			c.Header("Content-Type", "text/plain")
			return
		}
		c.String(200, fmt.Sprintf("%d", retrievedName))
	}
}
