package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/logger"
)

// update changes the value of global storage and returns a status code
func update(c *gin.Context, metricType string, db Database) {
	metricValue := c.Param("metricValue")
	metricName := c.Param("metricName")

	if metricValue == "" {
		logger.Log.Error("error while extracting metric params")
		c.String(http.StatusNotFound, "")
		return
	}

	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			logger.Log.Error("error while parsing float metric val")
			c.String(http.StatusBadRequest, "")
			return
		}
		db.Update("gauge", metricName, v)

	case "counter":
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			logger.Log.Error("error while parsing int metric val")
			c.String(http.StatusBadRequest, "")
			return
		}
		db.Update("counter", metricName, v)
	}
	c.String(http.StatusOK, "")
}

// Handler is a Gin route handler for POST HTTP metric updates
func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Param("metricType") != "gauge" && c.Param("metricType") != "counter" {
			logger.Log.Error("invalid metric type")
			c.String(http.StatusBadRequest, "")
			return
		}
		update(c, c.Param("metricType"), Storage)
	}
}

// return metric value from storage
func Return(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")

	switch metricType {
	case "gauge":
		retrievedName, ok := Storage.Gauge[metricName]
		if !ok {
			logger.Log.Error("no such metric")
			c.String(http.StatusNotFound, "")
			return
		}
		c.String(200, strconv.FormatFloat(retrievedName, 'g', -1, 64))

	case "counter":
		retrievedName, ok := Storage.Counter[metricName]
		if !ok {
			logger.Log.Error("no such metric")
			c.String(http.StatusNotFound, "")
			return
		}
		logger.Log.Sugar().Infof("sent response: %d", http.StatusOK)
		c.String(200, strconv.FormatInt(retrievedName, 10))

	default:
		logger.Log.Error("invalid metric type")
		c.String(http.StatusBadRequest, "")
	}
}

func ReturnAll(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"message": "Collected Metrics",
		"metrics": Storage,
	})
}
