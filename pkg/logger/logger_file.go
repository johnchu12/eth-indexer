package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cast"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitWithFile(name string) *zap.Logger {
	dir, _ := os.Getwd()
	targetLogFolder := fmt.Sprintf("%s/logs", dir)
	if _, err := os.Stat(targetLogFolder); os.IsNotExist(err) {
		if err := os.Mkdir(targetLogFolder, 0700); err != nil {
			panic(err)
		}
	}

	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.MessageKey = "msg"

	// cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	// cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// 	enc.AppendString(t.Format("2006-01-02T15:04:05"))
	// }
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(cast.ToString(t.UnixMilli()))
	}
	// cfg.OutputPaths = []string{} // default: stderr

	var cores []zapcore.Core

	lv := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s.log", targetLogFolder, name),
			MaxSize:    600,
			MaxBackups: 3,
			MaxAge:     3,
			Compress:   true,
			LocalTime:  true,
		}),
		lv,
	)
	cores = append(cores, core)

	lv2 := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	core2 := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s_error.log", targetLogFolder, name),
			MaxSize:    600,
			MaxBackups: 3,
			MaxAge:     3,
			Compress:   true,
			LocalTime:  true,
		}),
		lv2,
	)

	cores = append(cores, core2)

	options := []zap.Option{
		zap.AddCaller(),
		zap.Hooks(func(e zapcore.Entry) error {
			// log.Printf("%-5s| %s %+v", strings.ToUpper(e.Level.String()), e.Time.UTC().Format("2006-01-02T15:04:05"), e.Message)
			if e.Level > zapcore.DebugLevel {
				log.Printf("%-5s| %+v", strings.ToUpper(e.Level.String()), e.Message)
			}
			return nil
		}),
	}

	logger := zap.New(zapcore.NewTee(cores...), options...)

	// 會卡住
	// zap.RedirectStdLog(logger)

	zap.ReplaceGlobals(logger)

	return logger
}
