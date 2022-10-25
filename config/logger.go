package config

import "go.uber.org/zap"

func NewLogger(c *Config, opts ...zap.Option) (*zap.Logger, error) {
	lc := c.Log
	zapCfg := zap.NewProductionConfig()
	switch lc.Level {
	case "debug":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
	zapCfg.Development = lc.IsDevelopment
	zapCfg.OutputPaths = []string{lc.LogFile, "stdout"}
	zapCfg.ErrorOutputPaths = []string{lc.LogFile, "stdout"}

	return zapCfg.Build(opts...)
}
