package logging

import (
	"fmt"
	"sync"
)

var (
	defaultLogger     *Logger
	defaultLoggerOnce sync.Once
)

// InitDefaultLogger 初始化默认日志器（使用 sync.Once 保证只初始化一次）
func InitDefaultLogger(config LogConfig) error {
	var err error
	defaultLoggerOnce.Do(func() {
		defaultLogger, err = CreateLogger("default", config)
		if err != nil {
			fmt.Printf("init logger error: %v\n", err)
		}
	})
	return err
}

// ensureDefaultLogger 确保 defaultLogger 已初始化
func ensureDefaultLogger() {
	defaultLoggerOnce.Do(func() {
		if err := InitDefaultLogger(NewLogConfigEmpty()); err != nil {
			fmt.Printf("init logger error: %v\n", err)
		}
	})
}

// Sync 全局刷新函数
func Sync() error {
	ensureDefaultLogger()
	if defaultLogger != nil {
		return defaultLogger.Sync()
	}
	return nil
}

// 全局日志函数，直接转发到 default 日志器
func Debugf(template string, args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Debugf(template, args...)
	}
}

func Infof(template string, args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Infof(template, args...)
	}
}

func Warnf(template string, args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Warnf(template, args...)
	}
}

func Errorf(template string, args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Errorf(template, args...)
	}
}

func Fatalf(template string, args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Fatalf(template, args...)
	}
}

func Debug(args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Debug(args...)
	}
}

func Info(args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Info(args...)
	}
}

func Warn(args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Warn(args...)
	}
}

func Error(args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Error(args...)
	}
}

func Fatal(args ...interface{}) {
	ensureDefaultLogger()
	if defaultLogger != nil {
		defaultLogger.Fatal(args...)
	}
}
