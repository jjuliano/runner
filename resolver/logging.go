package resolver

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
)

var logger = log.Default()

// LogError logs an error message and exits the program
func LogError(message string, err error) {
	logger.Errorf("%s: %s", message, err)
	os.Exit(1)
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
