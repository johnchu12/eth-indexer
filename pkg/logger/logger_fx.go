package logger

import (
	"fmt"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// CustomZapLogger 實現了 fxevent.Logger 接口
type CustomZapLogger struct {
	Logger *zap.Logger
}

// LogEvent 處理不同類型的 Fx 事件
func (l *CustomZapLogger) LogEvent(event fxevent.Event) {
	var prefix string
	var msg string

	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		prefix = "[START]"
		msg = fmt.Sprintf("%s Function: %v", prefix, e.FunctionName)
	case *fxevent.OnStopExecuting:
		prefix = "[STOP]"
		msg = fmt.Sprintf("%s Function: %v", prefix, e.FunctionName)
	case *fxevent.Supplied:
		prefix = "[SUPPLY]"
		msg = fmt.Sprintf("%s Type: %v", prefix, e.TypeName)
	case *fxevent.Provided:
		prefix = "[PROVIDE]"
		msg = fmt.Sprintf("%s %v", prefix, e.ConstructorName)
	case *fxevent.Invoking:
		prefix = "[INVOKE]"
		msg = fmt.Sprintf("%s Function: %v", prefix, e.FunctionName)
	case *fxevent.Invoked:
		if e.Err != nil {
			prefix = "[INVOKED]"
			msg = fmt.Sprintf("%s Function: %v, error: %v", prefix, e.FunctionName, e.Err)
		} else {
			msg = "" // 不輸出訊息
		}
	case *fxevent.Run:
		prefix = "[RUN]"
		if e.Err != nil {
			msg = fmt.Sprintf("%s Run error: %v", prefix, e.Err)
		} else {
			msg = fmt.Sprintf("%s Run completed successfully.", prefix)
		}
	case *fxevent.LoggerInitialized:
		msg = "[LOGGER] Logger has been initialized."
		// default:
		// 	msg = fmt.Sprintf("[OTHER] %T Event: %v", event, event)
	}

	if msg != "" {
		switch event.(type) {
		case *fxevent.OnStartExecuting, *fxevent.OnStopExecuting, *fxevent.Supplied, *fxevent.Provided, *fxevent.Invoking, *fxevent.Invoked, *fxevent.Run, *fxevent.LoggerInitialized:
			if isErrorEvent(event) {
				l.Logger.Error(msg)
			} else {
				l.Logger.Debug(msg)
			}
		default:
			l.Logger.Error(msg)
		}
	}
}

// isErrorEvent 判斷事件是否為錯誤類型
func isErrorEvent(event fxevent.Event) bool {
	switch e := event.(type) {
	case *fxevent.Invoked:
		return e.Err != nil
	case *fxevent.Run:
		return e.Err != nil
	default:
		return false
	}
}
