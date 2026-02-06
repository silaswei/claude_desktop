package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

var (
	appLogger logger.Logger
	logDate  string // 当前日志文件的日期
)

// InitLogger 初始化日志系统（使用 Wails Logger）
func InitLogger() error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 创建日志目录
	logDir := filepath.Join(homeDir, ".claude-desktop", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 获取当前日期
	currentDate := time.Now().Format("2006-01-02")
	logDate = currentDate

	// 默认使用 app.log 作为日志文件名
	logPath := filepath.Join(logDir, "app.log")

	// 使用 Wails FileLogger
	appLogger = logger.NewFileLogger(logPath)

	// 记录启动信息
	appLogger.Info("========================================")
	appLogger.Info("Claude Desktop 应用启动")
	appLogger.Info(fmt.Sprintf("日志文件: %s", logPath))
	appLogger.Info(fmt.Sprintf("当前日期: %s", currentDate))
	appLogger.Info("========================================")

	return nil
}

// checkAndRotateLog 检查日期变化，必要时轮转日志文件
func checkAndRotateLog() {
	if appLogger == nil {
		return
	}

	currentDate := time.Now().Format("2006-01-02")

	// 如果日期变化，进行日志轮转
	if currentDate != logDate {
		homeDir, _ := os.UserHomeDir()
		logDir := filepath.Join(homeDir, ".claude-desktop", "logs")

		// 关闭当前日志（通过设置为 nil，将在重新初始化时创建新实例）
		appLogger.Info(fmt.Sprintf("日期变化，准备轮转日志: %s -> %s", logDate, currentDate))

		// 重命名当前日志文件：app.log -> app-YYYY-MM-DD.log
		oldLogPath := filepath.Join(logDir, "app.log")
		newLogPath := filepath.Join(logDir, fmt.Sprintf("app-%s.log", logDate))

		// 先关闭旧日志
		appLogger = nil

		// 重命名日志文件
		if err := os.Rename(oldLogPath, newLogPath); err != nil {
			// 如果重命名失败（可能文件不存在），创建一个新的 appLogger 记录错误
			appLogger = logger.NewFileLogger(oldLogPath)
			appLogger.Error(fmt.Sprintf("重命名日志文件失败: %v", err))
			return
		}

		// 重新初始化日志系统（会创建新的 app.log）
		if err := InitLogger(); err != nil {
			fmt.Printf("重新初始化日志系统失败: %v\n", err)
			return
		}

		// 记录日志轮转信息
		appLogger.Info(fmt.Sprintf("日志轮转完成: %s -> %s", oldLogPath, newLogPath))
	}
}

// CloseLogger 关闭日志系统
func CloseLogger() {
	if appLogger != nil {
		appLogger.Info("Claude Desktop 应用关闭")
	}
}

// Info 写入信息日志
func Info(format string, args ...interface{}) {
	checkAndRotateLog()
	if appLogger != nil {
		message := format
		if len(args) > 0 {
			message = fmt.Sprintf(format, args...)
		}
		appLogger.Info(message)
	}
}

// Error 写入错误日志
func Error(format string, args ...interface{}) {
	checkAndRotateLog()
	if appLogger != nil {
		message := format
		if len(args) > 0 {
			message = fmt.Sprintf(format, args...)
		}
		appLogger.Error(message)
	}
}

// Debug 写入调试日志
func Debug(format string, args ...interface{}) {
	checkAndRotateLog()
	if appLogger != nil {
		message := format
		if len(args) > 0 {
			message = fmt.Sprintf(format, args...)
		}
		appLogger.Debug(message)
	}
}

// Print 写入普通日志
func Print(format string, args ...interface{}) {
	checkAndRotateLog()
	if appLogger != nil {
		message := format
		if len(args) > 0 {
			message = fmt.Sprintf(format, args...)
		}
		appLogger.Print(message)
	}
}

// Warning 写入警告日志
func Warning(format string, args ...interface{}) {
	checkAndRotateLog()
	if appLogger != nil {
		message := format
		if len(args) > 0 {
			message = fmt.Sprintf(format, args...)
		}
		appLogger.Warning(message)
	}
}

// Trace 写入追踪日志
func Trace(format string, args ...interface{}) {
	checkAndRotateLog()
	if appLogger != nil {
		message := format
		if len(args) > 0 {
			message = fmt.Sprintf(format, args...)
		}
		appLogger.Trace(message)
	}
}

// FrontendLog 前端日志（通过后端调用）
func FrontendLog(message string) {
	checkAndRotateLog()
	if appLogger != nil {
		appLogger.Info(fmt.Sprintf("[FRONTEND] %s", message))
	}
}
