package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.SugaredLogger
	once     sync.Once
)

// Init 初始化日志 - 简洁 API
func Init(level, format string) {
	once.Do(func() {
		lvl := zapcore.DebugLevel
		_ = lvl.Set(level)

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "ts"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		var encoder zapcore.Encoder
		if format == "json" {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		ws := zapcore.AddSync(os.Stdout)
		core := zapcore.NewCore(encoder, ws, lvl)
		l := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		instance = l.Sugar()
	})
}

// L 获取底层 SugaredLogger
func L() *zap.SugaredLogger {
	if instance == nil {
		Init("debug", "console")
	}
	return instance
}

// Sync 刷新日志缓冲
func Sync() {
	if instance != nil {
		_ = instance.Sync()
	}
}

// ==================== 快捷函数 ====================

func Info(msg string, keysAndValues ...interface{}) {
	L().Infow(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	L().Errorw(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...interface{}) {
	L().Debugw(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	L().Warnw(msg, keysAndValues...)
}
