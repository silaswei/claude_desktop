package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile   *os.File
	logMutex  sync.Mutex
	logPrefix = ""
)

// InitLogger 初始化日志系统
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

	// 创建日志文件（按日期命名）
	dateStr := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("app-%s.log", dateStr))

	// 打开日志文件（追加模式）
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// 写入日志头部
	WriteLog("========================================")
	WriteLog("Claude Desktop 应用启动")
	WriteLog("日志文件: %s", logPath)
	WriteLog("========================================\n")

	return nil
}

// CloseLogger 关闭日志系统
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// WriteLog 写入日志
func WriteLog(format string, args ...interface{}) {
	if logFile == nil {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	// 生成时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// 写入日志
	fmt.Fprintf(logFile, "[%s] %s%s\n",
		timestamp,
		logPrefix,
		fmt.Sprintf(format, args...),
	)

	// 同时输出到控制台
	fmt.Printf("[%s] %s%s\n",
		timestamp,
		logPrefix,
		fmt.Sprintf(format, args...),
	)
}

// SetPrefix 设置日志前缀
func SetPrefix(prefix string) {
	logPrefix = "[" + prefix + "] "
}

// Info 写入信息日志
func Info(format string, args ...interface{}) {
	SetPrefix("INFO")
	WriteLog(format, args...)
	logPrefix = ""
}

// Error 写入错误日志
func Error(format string, args ...interface{}) {
	SetPrefix("ERROR")
	WriteLog(format, args...)
	logPrefix = ""
}

// Debug 写入调试日志
func Debug(format string, args ...interface{}) {
	SetPrefix("DEBUG")
	WriteLog(format, args...)
	logPrefix = ""
}

// FrontendLog 前端日志（通过后端调用）
func FrontendLog(message string) {
	SetPrefix("FRONTEND")
	WriteLog("%s", message)
	logPrefix = ""
}
