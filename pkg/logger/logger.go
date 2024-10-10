package logger

import (
	"os"
	"time"

	"github.com/golang-module/carbon/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Infow logs a message with the given key-value pairs.
func Infow(msg string, keysAndValues ...interface{}) {
	zap.S().Infow(msg, keysAndValues...)
}

// Infof logs a formatted message.
func Infof(template string, args ...interface{}) {
	zap.S().Infof(template, args...)
}

// Warnw logs a warning message with the given key-value pairs.
func Warnw(msg string, keysAndValues ...interface{}) {
	zap.S().Warnw(msg, keysAndValues...)
}

// Warnf logs a formatted warning message.
func Warnf(template string, args ...interface{}) {
	zap.S().Warnf(template, args...)
}

// Errorw logs an error message with the given key-value pairs and includes a stack trace.
func Errorw(msg string, keysAndValues ...interface{}) {
	stackTrace := zap.StackSkip("stacktrace", 1)
	zap.L().Sugar().Errorw(msg, append(keysAndValues, stackTrace)...)
}

// Errorf logs a formatted error message.
func Errorf(template string, args ...interface{}) {
	zap.S().Errorf(template, args...)
}

// ErrorfWithStack logs a formatted error message with an included stack trace.
func ErrorfWithStack(template string, args ...interface{}) {
	logger := zap.S().With(zap.StackSkip("stacktrace", 1))
	logger.Errorf(template, args...)
}

// InfoField logs an informational message with additional structured fields.
func InfoField(template string, fields ...zapcore.Field) {
	zap.L().Info(template, fields...)
}

// GetLogger returns the global zap.Logger instance.
func GetLogger() *zap.Logger {
	return zap.L()
}

// Init initializes the global logger with the specified configuration.
func Init() *zap.Logger {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(carbon.CreateFromTimestamp(t.Unix()).ToDateTimeString())
	}
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		os.Stderr,
		levelEnabler,
		// zapcore.AddSync(&lumberjack.Logger{
		// 	Filename:   fmt.Sprintf("%s/%s.log", targetLogFolder, name),
		// 	MaxSize:    600,
		// 	MaxBackups: 3,
		// 	MaxAge:     3,
		// 	Compress:   true,
		// 	LocalTime:  true,
		// }),
	)

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

	logger := zap.New(core, options...)

	zap.ReplaceGlobals(logger)

	return logger
}
