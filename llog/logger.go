package llog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log         *zap.Logger
	sugar       *zap.SugaredLogger
	initialized bool
)

// Init 初始化日志系统
func Init(logConfig *LogSetting) error {
	if initialized {
		return nil
	}

	if logConfig == nil {
		logConfig = &DefaultConfig
	}

	// 创建日志目录
	if logConfig.File {
		logDir := filepath.Clean(logConfig.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %v", err)
		}
	}

	// 设置日志级别
	logLevel := getLogLevel(logConfig.Level)

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:       "msg",
		LevelKey:         "level",
		TimeKey:          "time",
		NameKey:          "logger",
		CallerKey:        "caller",
		FunctionKey:      zapcore.OmitKey,
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "\t",
	}

	var cores []zapcore.Core

	// 添加控制台输出
	if logConfig.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel))
	}

	// 添加文件输出
	if logConfig.File {
		// 使用当前日期作为日志文件名
		fileName := fmt.Sprintf(logConfig.Format, time.Now().Format("2006-01-02"))
		path := filepath.Join(logConfig.FilePath, fileName)

		var fileEncoder zapcore.Encoder
		if logConfig.OutputFormat != "text" {
			fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Clean(path),
			MaxSize:    logConfig.MaxSize,
			MaxBackups: logConfig.MaxBackups,
			MaxAge:     logConfig.MaxAge,
			Compress:   logConfig.Compress,
			LocalTime:  logConfig.LocalTime,
		})
		cores = append(cores, zapcore.NewCore(fileEncoder, writer, logLevel))
	}

	// 创建核心
	core := zapcore.NewTee(cores...)

	// 创建logger
	log = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// 创建sugar logger
	sugar = log.Sugar()

	// 输出初始化信息
	sugar.Infof("日志系统初始化完成，日志级别: %s", strings.ToUpper(logConfig.Level))

	initialized = true
	return nil
}

// FormatError 格式化错误信息，去除重复
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	parts := strings.Split(err.Error(), ": ")
	seen := make(map[string]bool)
	var unique []string

	for _, part := range parts {
		if !seen[part] {
			seen[part] = true
			unique = append(unique, part)
		}
	}

	return strings.Join(unique, ": ")
}

// Debug 输出调试级别日志
func Debug(args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Debug(args...)
}

// Info 输出信息级别日志
func Info(args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Info(args...)
}

// Warn 输出警告级别日志
func Warn(args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Warn(args...)
}

// Error 输出错误级别日志
func Error(args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Error(args...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Fatal(args...)
}

func DebugF(format string, args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Debugf(format, args...)
}

func InfoF(format string, args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Infof(format, args...)
}

func WarnF(format string, args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Warnf(format, args...)
}

func ErrorF(format string, args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Errorf(format, args...)
}

func FatalF(format string, args ...interface{}) {
	if !initialized {
		return
	}
	sugar.Fatalf(format, args...)
}

// Sync 同步日志到磁盘
func Sync() {
	if !initialized || log == nil {
		return
	}

	_ = log.Sync()
	if sugar != nil {
		_ = sugar.Sync()
	}
}

// Cleanup 清理日志资源
func Cleanup() {
	if !initialized {
		return
	}

	if log != nil {
		Sync()
	}

	initialized = false
}

// HandlePanic 处理panic并记录日志
func HandlePanic() {
	if r := recover(); r != nil {
		stack := debug.Stack()
		errorMsg := fmt.Sprintf("程序发生严重错误: %v\n堆栈信息:\n%s", r, stack)
		Error(errorMsg)
		Sync()
		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	}
}
