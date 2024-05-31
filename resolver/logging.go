package resolver

import (
	"fmt"

	"github.com/charmbracelet/log"
)

var logger = log.Default()

// LogError logs an error message and returns the error
func LogError(message string, err error) error {
	msg := fmt.Sprintf("❌ %s: %s", message, err)
	logger.Errorf(msg) // log library does not return an error object

	return fmt.Errorf(msg)
}

// LogInfo logs an informational message
func LogInfo(message string) {
	logger.Info(message)
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

// GetLogger returns the logger instance
func GetLogger() *log.Logger {
	return logger
}
