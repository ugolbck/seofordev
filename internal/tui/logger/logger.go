package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	debugLogger *log.Logger
	logFile     *os.File
	debugMode   bool
)

// EnableDebugMode enables debug logging (for development)
func EnableDebugMode() {
	debugMode = true
}

// InitLogger initializes the debug logger to write to a file (only in debug mode)
func InitLogger() error {
	// Check for debug environment variable
	if os.Getenv("SEO_DEBUG") == "1" || os.Getenv("DEBUG") == "1" {
		debugMode = true
	}

	// Only initialize logging if debug mode is enabled
	if !debugMode {
		return nil
	}

	// Create logs directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".seo", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logDir, fmt.Sprintf("seofordev-%s.log", timestamp))

	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	debugLogger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)

	// Log startup
	debugLogger.Println("=== SEO CLI Debug Log Started ===")

	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if debugMode && logFile != nil {
		debugLogger.Println("=== SEO CLI Debug Log Ended ===")
		logFile.Close()
	}
}

// LogDebug logs a debug message (only in debug mode)
func LogDebug(format string, args ...interface{}) {
	if debugMode && debugLogger != nil {
		debugLogger.Printf("[DEBUG] "+format, args...)
	}
}

// LogInfo logs an info message (only in debug mode)
func LogInfo(format string, args ...interface{}) {
	if debugMode && debugLogger != nil {
		debugLogger.Printf("[INFO] "+format, args...)
	}
}

// LogError logs an error message (only in debug mode)
func LogError(format string, args ...interface{}) {
	if debugMode && debugLogger != nil {
		debugLogger.Printf("[ERROR] "+format, args...)
	}
}

// LogAPICall logs an API call (only in debug mode)
func LogAPICall(method, url string, err error) {
	if !debugMode {
		return
	}
	if err != nil {
		LogError("API %s %s failed: %v", method, url, err)
	} else {
		LogDebug("API %s %s succeeded", method, url)
	}
}

// LogExport logs export operations (only in debug mode)
func LogExport(operation string, pageCount int, err error) {
	if !debugMode {
		return
	}
	if err != nil {
		LogError("Export %s failed: %v", operation, err)
	} else {
		LogInfo("Export %s succeeded with %d pages", operation, pageCount)
	}
}

// LogUIEvent logs UI events for debugging (only in debug mode)
func LogUIEvent(component, event string, details ...interface{}) {
	if !debugMode {
		return
	}
	if len(details) > 0 {
		LogDebug("UI [%s] %s: %v", component, event, details)
	} else {
		LogDebug("UI [%s] %s", component, event)
	}
}
