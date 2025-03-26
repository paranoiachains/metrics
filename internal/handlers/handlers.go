package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/collector"
	"github.com/paranoiachains/metrics/internal/logger"
)

// update changes the value of global storage and returns a status code
func update(c *gin.Context, db Database) {
	var buf bytes.Buffer
	var metric collector.Metric

	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		logger.Log.Error("error while reading from request body")
		c.String(http.StatusInternalServerError, "")
	}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		logger.Log.Error("error while decoding json")
		c.String(http.StatusInternalServerError, "")
	}
	switch metric.MType {
	case "gauge":
		db.Update(metric.MType, metric.ID, *metric.Value)
	case "counter":
		db.Update(metric.MType, metric.ID, *metric.Delta)
	}

	resp, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error("error while encoding json")
		c.String(http.StatusInternalServerError, "")
	}
	c.JSON(http.StatusOK, resp)
}

// Handler is a Gin route handler for POST HTTP metric updates
func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		update(c, Storage)
	}
}

// return metric value from storage
func Return(c *gin.Context, db Database) {
	var buf bytes.Buffer
	var reqMetric collector.Metric

	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		logger.Log.Error("error while reading from request body")
		c.String(http.StatusInternalServerError, "")
	}
	if err = json.Unmarshal(buf.Bytes(), &reqMetric); err != nil {
		logger.Log.Error("error while decoding json")
		c.String(http.StatusInternalServerError, "")
	}
	respMetric, err := db.Return(reqMetric.MType, reqMetric.ID)
	if err != nil {
		logger.Log.Error("error while getting metric from db")
	}

	resp, err := json.Marshal(respMetric)
	if err != nil {
		logger.Log.Error("error while encoding json")
		c.String(http.StatusInternalServerError, "")
	}
	c.JSON(http.StatusOK, resp)
}

func ReturnWrap() gin.HandlerFunc {
	return func(c *gin.Context) {
		Return(c, Storage)
	}
}

func ReturnAll(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"message": "Collected Metrics",
		"metrics": Storage,
	})
}
