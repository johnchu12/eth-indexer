package logger

import (
	"os"
	"strings"
	"time"

	"github.com/golang-module/carbon/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// gin logger
type CronLog struct {
	Logger interface {
		// Info logs routine messages about cron's operation.
		Info(msg string, keysAndValues ...interface{})
		// Error logs an error condition.
		Error(err error, msg string, keysAndValues ...interface{})
	}
}

func (l *CronLog) Info(msg string, keysAndValues ...interface{}) {
	zap.S().Infow(msg, keysAndValues...)
}

func (l *CronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	zap.S().Warnw(msg, keysAndValues...)
}

func (l *CronLog) Write(p []byte) (n int, err error) {
	l.Logger.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

// global logger
func Infow(msg string, keysAndValues ...interface{}) {
	zap.S().Infow(msg, keysAndValues...)
}
func Infof(template string, args ...interface{}) {
	zap.S().Infof(template, args...)
}
func Warnw(msg string, keysAndValues ...interface{}) {
	zap.S().Warnw(msg, keysAndValues...)
}
func Warnf(template string, args ...interface{}) {
	zap.S().Warnf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	// 指定跳過的調用棧幀數,這裡跳過了 logger 包內部的調用
	stackTrace := zap.StackSkip("stacktrace", 1)

	// 記錄錯誤日誌,並包含調用棧信息
	zap.L().Sugar().Errorw(msg, append(keysAndValues, stackTrace)...)
}

func Errorf(template string, args ...interface{}) {
	// 使用修改後的 logger 記錄錯誤
	zap.S().Errorf(template, args...)
}

func ErrorfWithStack(template string, args ...interface{}) {
	// 使用 zap.S().With() 添加堆棧跟蹤作為結構化字段
	logger := zap.S().With(zap.StackSkip("stacktrace", 1))

	// 使用修改後的 logger 記錄錯誤
	logger.Errorf(template, args...)
}

func InfoField(template string, fields ...zapcore.Field) {
	zap.L().Info(template, fields...)
}

func GetLogger() *zap.Logger {
	return zap.L()
}

func Init() *zap.Logger {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		// enc.AppendString(cast.ToString(t.UnixMilli()))
		enc.AppendString(carbon.CreateFromTimestamp(t.Unix()).ToDateTimeString())
	}
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	var cores []zapcore.Core

	lv := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		os.Stderr,
		// zapcore.AddSync(&lumberjack.Logger{
		// 	Filename:   fmt.Sprintf("%s/%s.log", targetLogFolder, name),
		// 	MaxSize:    600,
		// 	MaxBackups: 3,
		// 	MaxAge:     3,
		// 	Compress:   true,
		// 	LocalTime:  true,
		// }),
		lv,
	)
	cores = append(cores, core)

	options := []zap.Option{
		// zap.AddCaller(),
		// zap.Hooks(func(e zapcore.Entry) error {
		// 	// log.Printf("%-5s| %s %+v", strings.ToUpper(e.Level.String()), e.Time.UTC().Format("2006-01-02T15:04:05"), e.Message)
		// 	if e.Level > zapcore.DebugLevel {
		// 		log.Printf("%-5s| %+v", strings.ToUpper(e.Level.String()), e.Message)
		// 	}
		// 	return nil
		// }),
	}

	logger := zap.New(zapcore.NewTee(cores...), options...)

	// 會卡住
	// zap.RedirectStdLog(logger)

	zap.ReplaceGlobals(logger)

	return logger
}
