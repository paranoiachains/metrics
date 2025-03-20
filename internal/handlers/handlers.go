package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/utils"
)

// updateMetric changes the value of global storage and returns a status code
func updateMetric(c *gin.Context, metricType string, db Database) {
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
		db.Update("gauge", metricName, v)

	case "counter":
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.Status(http.StatusBadRequest)
			c.Header("Content-Type", "text/plain")
			c.Header("Error", err.Error())
			return
		}
		db.Update("counter", metricName, v)
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
		updateMetric(c, c.Param("metricType"), Storage)
		c.Header("Content-Type", "text/plain")
	}
}

// return metric value from storage
func ReturnMetric(c *gin.Context) {
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
		c.String(200, strconv.FormatFloat(retrievedName, 'g', -1, 64))

	case "counter":
		retrievedName, ok := Storage.Counter[metricName]
		if !ok {
			c.Status(http.StatusNotFound)
			c.Header("Content-Type", "text/plain")
			return
		}
		c.String(200, strconv.FormatInt(retrievedName, 10))

	default:
		c.String(http.StatusBadRequest, "Invalid metric type")
	}
}

func ReturnAll(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"message": "Collected Metrics",
		"metrics": Storage,
	})
}
