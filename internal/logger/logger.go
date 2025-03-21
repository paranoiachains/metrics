package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Initialize() error {
	cfg := zap.NewProductionConfig()
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = logger
	return nil
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)

		Log.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Duration("duration", duration),
		)
		Log.Info("HTTP Response",
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()))
	}
}
