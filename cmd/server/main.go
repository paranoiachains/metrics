package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/flags"
	"github.com/paranoiachains/metrics/internal/handlers"
	"github.com/paranoiachains/metrics/internal/logger"
	"github.com/paranoiachains/metrics/internal/middleware"
	"github.com/paranoiachains/metrics/internal/storage"
	"go.uber.org/zap"
)

func main() {
	logger.Initialize()

	flags.ParseServerFlags()
	logger.Log.Info("flags",
		zap.Bool("Restore?", flags.Restore),
		zap.String("Path", flags.FileStoragePath),
		zap.Int("Store interval", flags.StoreInterval))

	os.Mkdir("../../tmp", 0666)
	if !flags.Restore {
		storage.Storage.Clear()
		_, err := os.Create(flags.FileStoragePath)
		if err != nil {
			logger.Log.Error("error", zap.Error(err))
		}
	} else {
		storage.Storage.Restore(flags.FileStoragePath)
	}

	if flags.Cfg.Address != "" {
		flags.ServerEndpoint = flags.Cfg.Address
	}

	go storage.WriteWithInterval(storage.Storage, flags.FileStoragePath, flags.StoreInterval)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.LoggerMiddleware(), middleware.GzipMiddleware())

	// HTML response
	r.GET("/", handlers.HTMLReturnAll)

	// JSON requests
	r.POST("/update/", handlers.JSONUpdate())
	r.POST("/value/", handlers.JSONValue())

	// casual url requests
	r.POST("/update/:metricType/:metricName/:metricValue", handlers.URLUpdate())
	r.GET("/value/:metricType/:metricName/", handlers.URLValue)

	r.Run(flags.ServerEndpoint)
}
