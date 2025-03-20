package main

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/flags"
	"github.com/paranoiachains/metrics/internal/handlers"
)

func main() {
	flags.ParseServerFlags()
	// clear storage before init
	handlers.Storage.Clear()

	r := gin.Default()
	templatesPath, _ := filepath.Abs("../../templates/index.html")
	r.LoadHTMLFiles(templatesPath)

	r.POST("/update/:metricType/:metricName/:metricValue/", handlers.MetricHandler())
	r.GET("/value/:metricType/:metricName/", handlers.ReturnMetric)
	r.GET("/", handlers.ReturnAll)
	r.Run(flags.ServerEndpoint)
}
