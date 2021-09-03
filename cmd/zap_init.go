package cmd

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initLog(logLevel string) *zap.Logger {
	level := zapcore.InfoLevel
	if err := level.Set(logLevel); err != nil {
		log.Fatalf("could not set zap log level to: \"%s\" \n", logLevel)
	}

	config := &zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "severity",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.RFC3339NanoTimeEncoder,

			CallerKey:      "src",
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,

			StacktraceKey: "stk",
		},
	}

	zl, err := config.Build()
	if err != nil {
		log.Fatalf("failed to build zap logger. Error: %s", err)
	}
	return zl

}
