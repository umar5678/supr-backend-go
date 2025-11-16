package logger

import (
	"github.com/umar5678/go-backend/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Initialize sets up the logger
func Initialize(cfg *config.LoggerConfig) error {
	var zapConfig zap.Config

	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Configure output
	if cfg.Output == "file" && cfg.FilePath != "" {
		zapConfig.OutputPaths = []string{cfg.FilePath}
		zapConfig.ErrorOutputPaths = []string{cfg.FilePath}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	log = logger
	return nil
}

// Get returns the logger instance
func Get() *zap.Logger {
	if log == nil {
		// Fallback logger
		log, _ = zap.NewProduction()
	}
	return log
}

// Helper functions for common log levels
func Debug(msg string, fields ...interface{}) {
	Get().Sugar().Debugw(msg, fields...)
}

func Info(msg string, fields ...interface{}) {
	Get().Sugar().Infow(msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
	Get().Sugar().Warnw(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	Get().Sugar().Errorw(msg, fields...)
}

func Fatal(msg string, fields ...interface{}) {
	Get().Sugar().Fatalw(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	if log != nil {
		log.Sync()
	}
}
