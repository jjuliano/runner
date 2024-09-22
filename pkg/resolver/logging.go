package resolver

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
)

var (
	logger  = log.Default()
	verbose = os.Getenv("VERBOSE")
)

func shouldLog() bool {
	return verbose != ""
}

func LogErrorExit(message string, err error) {
	if shouldLog() {
		msg := fmt.Sprintf("❌ %s: %s", message, err)
		logger.Errorf(msg)
	}
	os.Exit(1)
}

func LogError(message string, err error) error {
	if shouldLog() {
		msg := fmt.Sprintf("❌ %s: %s", message, err)
		logger.Errorf(msg)
		return fmt.Errorf(msg)
	}
	return err
}

func LogInfo(message string) {
	logger.Info(message)
}

func LogDebug(message string) {
	if shouldLog() {
		logger.Debug(message)
	}
}

func PrintMessage(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func Println(a ...interface{}) {
	fmt.Println(a...)
}

func PrintError(message string, err error) {
	fmt.Printf("%s: %v\n", message, err)
}

func LogWarn(message string) {
	if shouldLog() {
		logger.Warn(message)
	}
}

func GetLogger() *log.Logger {
	return logger
}
