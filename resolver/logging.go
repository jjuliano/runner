package resolver

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
)

var logger = log.Default()

// LogErrorExit logs an error message and exits the program
func LogErrorExit(message string, err error) {
	msg := fmt.Sprintf("❌ %s: %s", message, err)
	logger.Errorf(msg)
	os.Exit(1)
}

// LogError logs an error message and returns an error
func LogError(message string, err error) error {
	msg := fmt.Sprintf("❌ %s: %s", message, err)
	logger.Errorf(msg)
	return fmt.Errorf(msg)
}

// LogInfo logs an informational message
func LogInfo(message string) {
	logger.Info(message)
}

// LogDebug logs a debugging message
func LogDebug(message string) {
	logger.Warn(message)
}

// PrintMessage prints a formatted message to the standard output
func PrintMessage(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// Println prints a message followed by a newline to the standard output
func Println(a ...interface{}) {
	fmt.Println(a...)
}

// PrintError prints an error message to the standard output
func PrintError(message string, err error) {
	fmt.Printf("%s: %v\n", message, err)
}

// LogWarn prints a warning message to the standard output
func LogWarn(message string) {
	logger.Warn(message)
}

// GetLogger returns the logger instance
func GetLogger() *log.Logger {
	return logger
}
