package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"p2p_remote_desk/config"
	"path/filepath"
	"strings"
	"time"
)

var (
	log         *zap.Logger
	sugar       *zap.SugaredLogger
	initialized bool
)

// Init 初始化日志系统
func Init() error {
	if initialized {
		return nil
	}

	// 创建日志目录
	logConfig := config.GetConfig().LogConfig
	logDir := filepath.Join(logConfig.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 使用当前日期作为日志文件名
	fileName := fmt.Sprintf(logConfig.Format, time.Now().Format("2006-01-02"))
	path := filepath.Join(logDir, fileName)

	// 配置日志轮转
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path,
		MaxSize:    logConfig.MaxSize,
		MaxBackups: logConfig.MaxBackups,
		MaxAge:     logConfig.MaxAge,
		Compress:   logConfig.Compress,
		LocalTime:  logConfig.LocalTime,
	})

	// 控制台编码器配置
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 文件编码器配置
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建编码器
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	fileEncoder := zapcore.NewConsoleEncoder(fileEncoderConfig)

	// 根据环境设置日志级别
	var logLevel zapcore.Level
	if config.IsDevelopment() {
		logLevel = zapcore.DebugLevel
	} else {
		logLevel = zapcore.InfoLevel
	}

	// 创建核心，同时输出到控制台和文件
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel),
		zapcore.NewCore(fileEncoder, writer, logLevel),
	)

	// 创建logger
	log = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// 创建sugar logger用于格式化输出
	sugar = log.Sugar()

	// 输出当前日志级别
	level := "INFO"
	if logLevel == zapcore.DebugLevel {
		level = "DEBUG"
	}
	sugar.Infof("日志系统初始化完成，当前环境: %s，日志级别: %s", config.GetConfig().Environment, level)

	initialized = true
	return nil
}

// formatError 格式化错误信息,去除重复
func formatError(err error) string {
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

// Debug 记录调试级别日志
func Debug(format string, v ...interface{}) {
	if !initialized {
		if err := Init(); err != nil {
			fmt.Printf("[INIT_ERROR] %v\n", err)
			return
		}
	}

	msg := formatMessage(format, v...)
	sugar.Debug(msg)
}

// Info 记录信息级别日志
func Info(format string, v ...interface{}) {
	if !initialized {
		if err := Init(); err != nil {
			fmt.Printf("[INIT_ERROR] %v\n", err)
			return
		}
	}

	msg := formatMessage(format, v...)
	sugar.Info(msg)
}

// Warn 记录警告级别日志
func Warn(format string, v ...interface{}) {
	if !initialized {
		if err := Init(); err != nil {
			fmt.Printf("[INIT_ERROR] %v\n", err)
			return
		}
	}

	msg := formatMessage(format, v...)
	sugar.Warn(msg)
}

// Error 记录错误级别日志
func Error(format string, v ...interface{}) {
	if !initialized {
		if err := Init(); err != nil {
			fmt.Printf("[INIT_ERROR] %v\n", err)
			return
		}
	}

	msg := formatMessage(format, v...)
	sugar.Error(msg)
}

// formatMessage 格式化消息
func formatMessage(format string, v ...interface{}) string {
	if len(v) == 1 {
		if err, ok := v[0].(error); ok {
			return formatError(err)
		}
	} else if len(v) > 0 {
		if err, ok := v[len(v)-1].(error); ok {
			v[len(v)-1] = formatError(err)
		}
	}
	return fmt.Sprintf(format, v...)
}

// Cleanup 清理日志资源
func Cleanup() {
	if !initialized {
		return
	}

	if log != nil {
		_ = log.Sync()
	}

	initialized = false
}
