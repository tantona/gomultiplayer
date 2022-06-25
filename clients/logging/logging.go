package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(name string) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	var logger, _ = cfg.Build()
	return logger.Named(name)
}
