package logger

import (
	"fmt"

	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

// EchoLogger 结构体，用于 Echo 日志记录
type EchoLogger struct {
	zapLogger *zap.Logger
}

// 实现 Echo Logger 接口的方法
func (l *EchoLogger) Info(i ...interface{}) {
	l.zapLogger.Info(fmt.Sprint(i...))
}
func (l *EchoLogger) Infof(template string, args ...interface{}) {
	l.zapLogger.Sugar().Infof(template, args...)
}
func (l *EchoLogger) Infoj(j log.JSON) {
	l.zapLogger.Info("", zap.Any("data", j))
}

func (l *EchoLogger) Warn(i ...interface{}) {
	l.zapLogger.Warn(fmt.Sprint(i...))
}
func (l *EchoLogger) Warnf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Warnf(template, args...)
}
func (l *EchoLogger) Warnj(j log.JSON) {
	l.zapLogger.Warn("", zap.Any("data", j))
}

func (l *EchoLogger) Debug(i ...interface{}) {
	l.zapLogger.Debug(fmt.Sprint(i...))
}
func (l *EchoLogger) Debugf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Debugf(format, args...)
}
func (l *EchoLogger) Debugj(j log.JSON) {
	l.zapLogger.Debug("", zap.Any("data", j))
}

func (l *EchoLogger) Error(i ...interface{}) {
	l.zapLogger.Error(fmt.Sprint(i...))
}
func (l *EchoLogger) Errorf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Errorf(format, args...)
}
func (l *EchoLogger) Errorj(j log.JSON) {
	l.zapLogger.Error("", zap.Any("data", j))
}

func (l *EchoLogger) Fatal(i ...interface{}) {
	l.zapLogger.Fatal(fmt.Sprint(i...))
}
func (l *EchoLogger) Fatalf(format string, args ...interface{}) {
	l.zapLogger.Sugar().Fatalf(format, args...)
}
func (l *EchoLogger) Fatalj(j log.JSON) {
	l.zapLogger.Fatal("", zap.Any("data", j))
}

func InitEchoLogger() *EchoLogger {
	zapLogger := Init() // 使用 Init 函数初始化 zap.Logger
	return &EchoLogger{
		zapLogger: zapLogger,
	}
}
