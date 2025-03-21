package main

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/flags"
	"github.com/paranoiachains/metrics/internal/handlers"
	"github.com/paranoiachains/metrics/internal/logger"
)

func main() {
	flags.ParseServerFlags()
	flags.ParseEnv()
	if flags.Cfg.Address != "" {
		flags.ServerEndpoint = flags.Cfg.Address
	}
	// clear storage before init
	handlers.Storage.Clear()

	r := gin.New()
	r.Use(gin.Recovery(), logger.Middleware())

	templatesPath, _ := filepath.Abs("../../templates/index.html")
	r.LoadHTMLFiles(templatesPath)

	r.POST("/update/:metricType/:metricName/:metricValue", handlers.Handler())
	r.GET("/value/:metricType/:metricName/", handlers.Return)
	r.GET("/", handlers.ReturnAll)
	r.Run(flags.ServerEndpoint)
}
