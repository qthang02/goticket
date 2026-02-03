package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Log *zap.Logger

func InitLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewJSONEncoder(config)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
	)

	Log = zap.New(core, zap.AddCaller())
}

func Info(message string, fields ...zap.Field) {
	Log.Info(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	Log.Error(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	Log.Debug(message, fields...)
}
