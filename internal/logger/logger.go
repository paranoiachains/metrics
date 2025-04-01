package logger

import (
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
