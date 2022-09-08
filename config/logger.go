package config

import (
	"fmt"
	"path"
	"time"

	zaprotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var level zapcore.Level

func NewLogger(c *Config) (logger *zap.Logger) {
	lc := c.Log

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
		logger = zap.New(getEncoderCore(), zap.AddStacktrace(level))
	} else {
		logger = zap.New(getEncoderCore())
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

// getEncoderCore 获取 Encoder 的 zapcore.Core
func getEncoderCore() (core zapcore.Core) {
	writer, err := getWriteSyncer()
	if err != nil {
		fmt.Printf("Get Write Syncer Failed err:%v \n", err.Error())
		return
	}
	return zapcore.NewCore(getEncoder(), writer, level)
}

// CustomTimeEncoder .
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("starnet_chain_api:" + "2006/01/02 - 15:04:05.000"))
}

func getWriteSyncer() (zapcore.WriteSyncer, error) {
	// file rotate logs split log
	fileWriter, err := zaprotatelogs.New(
		path.Join("log", "starnet_chain_api-%Y-%m-%d.log"),
		zaprotatelogs.WithRotationTime(24*time.Hour),
	)
	// return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter)), err
}
