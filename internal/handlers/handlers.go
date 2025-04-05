package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paranoiachains/metrics/internal/collector"
	"github.com/paranoiachains/metrics/internal/flags"
	"github.com/paranoiachains/metrics/internal/logger"
	"github.com/paranoiachains/metrics/internal/storage"
	"go.uber.org/zap"
)

func urlHandle(c *gin.Context, metricType string, db storage.Database) {
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

// URLUpdate is a Gin route handler for POST HTTP metric updates
func URLUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Param("metricType") != "gauge" && c.Param("metricType") != "counter" {
			logger.Log.Error("invalid metric type")
			c.String(http.StatusBadRequest, "")
			return
		}
		urlHandle(c, c.Param("metricType"), storage.Storage)
	}
}

// return metric value from storage
func URLValue(c *gin.Context) {
	metricType := c.Param("metricType")
	metricName := c.Param("metricName")

	switch metricType {
	case "gauge":
		retrievedName, ok := storage.Storage.Gauge[metricName]
		if !ok {
			logger.Log.Error("no such metric")
			c.String(http.StatusNotFound, "")
			return
		}
		c.String(200, strconv.FormatFloat(retrievedName, 'g', -1, 64))

	case "counter":
		retrievedName, ok := storage.Storage.Counter[metricName]
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

// jsonHandle changes the value of global storage and returns a status code
func jsonHandle(c *gin.Context, db storage.Database) {
	var buf bytes.Buffer
	var metric collector.Metric

	_, err := buf.ReadFrom(c.Request.Body)
	logger.Log.Info("request body:", zap.ByteString("body", buf.Bytes()))
	if err != nil {
		logger.Log.Error("error while reading from request body")
		c.String(http.StatusNotFound, "")
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		logger.Log.Error("error while decoding json", zap.Error(err))
		c.String(http.StatusNotFound, "")
		return
	}
	if metric.ID == "" {
		logger.Log.Error("metric id not found", zap.String("metric id", metric.ID))
		c.String(http.StatusNotFound, "")
		return
	}
	if metric.Delta == nil && metric.Value == nil {
		logger.Log.Error("no metric value!")
		c.String(http.StatusBadRequest, "")
		return
	}
	switch metric.MType {
	case "gauge":
		db.Update(metric.MType, metric.ID, *metric.Value)
	case "counter":
		db.Update(metric.MType, metric.ID, *metric.Delta)
	default:
		c.String(http.StatusBadRequest, "")
	}

	c.JSON(http.StatusOK, metric)
}

// JSONUpdate is a Gin route handler for POST HTTP metric updates
func JSONUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		jsonHandle(c, storage.Storage)
	}
}

// return metric returnValue from storage
func returnValue(c *gin.Context, db storage.Database) {
	var buf bytes.Buffer
	var reqMetric collector.Metric

	_, err := buf.ReadFrom(c.Request.Body)
	logger.Log.Info("request body:", zap.ByteString("body", buf.Bytes()))
	if err != nil {
		logger.Log.Error("error while reading from request body")
		c.String(http.StatusInternalServerError, "")
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &reqMetric); err != nil {
		logger.Log.Error("error while decoding json")
		c.String(http.StatusBadRequest, "")
		return
	}
	logger.Log.Info("unmarshalled metric:", zap.Object("metric", reqMetric))
	// debugging
	if strings.Contains(reqMetric.ID, "GetSet") {
		logger.Log.Info("current state of db", zap.Any("storage", storage.Storage))
	}
	if reqMetric.ID == "" {
		logger.Log.Error("metric id not found", zap.String("metric id", reqMetric.ID))
		c.String(http.StatusNotFound, "")
		return
	}
	respMetric, err := db.Return(reqMetric.MType, reqMetric.ID)
	if err != nil {
		logger.Log.Error("error while getting metric from db", zap.Error(err))
		c.String(http.StatusNotFound, "")
		return
	}
	c.JSON(http.StatusOK, respMetric)
}

func JSONValue() gin.HandlerFunc {
	return func(c *gin.Context) {
		returnValue(c, storage.Storage)
	}
}

func HTMLReturnAll(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK,
		`<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
		</head>
		<body>
			<h1>{{ .message }}</h1>
			<p>{{ .metrics }}</p>
		</body>
		</html>`)
}

func Ping(c *gin.Context) {
	flags.ParseServerFlags()

	host := flags.DBEndpoint
	user := flags.Cfg.DBUser
	password := flags.Cfg.DBPassword
	name := flags.Cfg.DBName
	dataSourceName := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, name)

	_, err := storage.ConnectAndPing("pgx", dataSourceName)
	if err != nil {
		logger.Log.Error("error while connecting to db", zap.Error(err))
		c.String(http.StatusInternalServerError, "")
		return
	}
	c.String(http.StatusOK, "pong")
}
