package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"test-golang-project/config"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger initializes rotating file logging configured via .env
func InitLogger(cfg *config.Config) io.Writer {
	logsDir := cfg.Log.Dir
	if logsDir == "" {
		logsDir = filepath.Join("storage", "logs")
	}

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("Warning: failed to create log directory: %v\n", err)
	}

	currentDate := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(logsDir, fmt.Sprintf("app-%s.log", currentDate))

	// Lumberjack handles rotation, retention days, and compression
	logFile := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}

	// Write simultaneously to terminal (stdout) and log file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Determine log level from .env
	var logLevel slog.Level
	switch strings.ToLower(cfg.Log.Level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handlerOpts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if cfg.AppEnv == "development" {
		handler = slog.NewTextHandler(multiWriter, handlerOpts)
	} else {
		handler = slog.NewJSONHandler(multiWriter, handlerOpts)
	}

	slog.SetDefault(slog.New(handler))

	// Forward Gin HTTP request and error logs to the same writer
	gin.DefaultWriter = multiWriter
	gin.DefaultErrorWriter = multiWriter

	slog.Info("Logger initialized", "file", logFilePath, "level", cfg.Log.Level, "env", cfg.AppEnv)
	return multiWriter
}
