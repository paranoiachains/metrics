package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ignoring gin default logger for the sake of learning

var Log *zap.Logger = zap.NewNop()

func Initialize() error {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig = zapcore.EncoderConfig{
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = logger
	return nil
}

func Middleware() gin.HandlerFunc {
	Initialize()
	defer Log.Sync()
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
			zap.Int("size", c.Writer.Size()),
			zap.String("Content-Type", c.Writer.Header().Get("Content-Type")),
		)
	}
}
