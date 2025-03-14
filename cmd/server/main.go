package main

import (
	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/handlers"
)

func main() {
	// clear storage before init
	handlers.Storage.Clear()

	r := gin.Default()
	r.POST("/update/:metricType/:metricName/:metricValue/", handlers.MetricHandler())
	r.Run()
}
