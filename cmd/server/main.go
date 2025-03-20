package main

import (
	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/flags"
	"github.com/paranoiachains/metrics/internal/handlers"
)

func main() {
	flags.ParseServerFlags()
	// clear storage before init
	handlers.Storage.Clear()

	r := gin.Default()
	r.POST("/update/:metricType/:metricName/:metricValue/", handlers.MetricHandler())
	r.GET("/value/:metricType/:metricName/", handlers.ReturnMetric)
	r.Run(flags.ServerEndpoint)
}
