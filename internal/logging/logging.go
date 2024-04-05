package logging

import (
	"go.uber.org/zap/zapcore"
	"os"

	"go.uber.org/zap"
)

var logger *zap.Logger

func Setup(region, serviceName, serviceVersion string) {

	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(getLogLevel()),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "timestamp",
			NameKey:        "logger",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			SkipLineEnding: false,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	var err error

	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}

	logger = logger.With(
		zap.String("region", region),
		zap.String("service", serviceName),
		zap.String("version", serviceVersion),
	)
}

func getLogger() *zap.Logger {
	return logger
}

func getLogLevel() zapcore.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
