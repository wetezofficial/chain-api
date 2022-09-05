package config

import (
	"fmt"
	"path"
	"time"

	zaprotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(c *Config) (logger *zap.Logger) {
	lc := c.Log
	var level zapcore.Level

	switch lc.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	if level == zap.DebugLevel || level == zap.ErrorLevel {
		logger = zap.New(getEncoderCore(level), zap.AddStacktrace(level))
	} else {
		logger = zap.New(getEncoderCore(level))
	}
	logger = logger.WithOptions(zap.AddCaller())

	return logger
}

// getEncoderConfig 获取 zapcore.EncoderConfig
func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	return config
}

// getEncoder 获取 zap core.Encoder
func getEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

// getEncoderCore 获取Encoder的 zap core.Core
func getEncoderCore(lvl zapcore.LevelEnabler) (core zapcore.Core) {
	writer, err := getWriteSyncer()
	if err != nil {
		fmt.Printf("Get Write Syncer Failed err:%v \n", err.Error())
		return
	}
	return zapcore.NewCore(getEncoder(), writer, lvl)
}

// CustomTimeEncoder .
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("chain-api:" + "2006/01/02 - 15:04:05.000"))
}

func getWriteSyncer() (zapcore.WriteSyncer, error) {
	// file-rotatelogs split log
	fileWriter, err := zaprotatelogs.New(
		path.Join("log", "%Y-%m-%d.log"),
		zaprotatelogs.WithRotationTime(24*time.Hour),
	)
	// return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter)), err
}
