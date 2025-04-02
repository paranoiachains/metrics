package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/flags"
	"github.com/paranoiachains/metrics/internal/handlers"
	"github.com/paranoiachains/metrics/internal/middleware"
	"github.com/paranoiachains/metrics/internal/storage"
)

func main() {
	flags.ParseServerFlags()
	if !flags.Restore {
		storage.Storage.Clear()
		file, err := os.Create(flags.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	} else {
		storage.Storage.Restore(flags.FileStoragePath)
	}

	if flags.Cfg.Address != "" {
		flags.ServerEndpoint = flags.Cfg.Address
	}

	fmt.Println("asd")
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
	fmt.Println("asd")
	go storage.WriteWithInterval(storage.Storage, flags.Cfg.FileStoragePath, flags.StoreInterval)
}
