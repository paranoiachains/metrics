package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrMetricVal = errors.New("extract_metric_params: metric val error")
	ErrNoName    = errors.New("extract_metric_params: no metric name")
)

func ExtractMetricParams(c *gin.Context) (string, string, error) {
	metricValue := c.Param("metricValue")
	metricName := c.Param("metricName")

	if metricValue == "" {
		return "", "", ErrMetricVal
	}
	if metricName == "" {
		return "", "", ErrNoName
	}
	return metricValue, metricName, nil
}
