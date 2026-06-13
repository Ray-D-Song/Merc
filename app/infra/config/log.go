package config

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger builds a zap.Logger configured from AppConfig.
func NewLogger(cfg *AppConfig) (*zap.Logger, error) {
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.TimeKey = "ts"
	zapCfg.EncoderConfig.CallerKey = "caller"

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}

	return logger.With(
		zap.String("server_port", cfg.Server.Port),
	), nil
}
