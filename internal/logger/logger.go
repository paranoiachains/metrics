package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ignoring gin default logger for the sake of learning

var Log *zap.Logger = zap.NewNop()

func Initialize() error {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	cfg.EncoderConfig = zapcore.EncoderConfig{
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
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
