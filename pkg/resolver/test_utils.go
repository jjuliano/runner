package resolver

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// captureOutput captures the standard output for testing
func captureOutput(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		PrintError("Error creating pipe", err)
		return ""
	}

	old := os.Stdout
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			PrintError("Error copying output", err)
		}
		outC <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = old
	return <-outC
}

func createWorkDir() string {
	tmpDir, err := os.MkdirTemp("", "runner_workdir")
	if err != nil {
		log.Fatalf("Failed to create work directory: %v", err)
	}
	return tmpDir
}

func writeEnvToFile(envFilePath string) error {
	envFile, err := os.Create(envFilePath)
	if err != nil {
		return fmt.Errorf("error creating env file: %w", err)
	}
	envFile.Sync()
	defer envFile.Close()

	if err = os.Setenv("RUNNER_ENV", envFilePath); err != nil {
		return err
	}

	for _, env := range os.Environ() {
		keyValue := strings.SplitN(env, "=", 2)
		key, value := keyValue[0], keyValue[1]

		if strings.ContainsAny(value, " \t\n\r\"'") {
			value = strconv.Quote(value)
		}

		if _, err := envFile.WriteString(fmt.Sprintf("%s=%s\n", key, value)); err != nil {
			return err
		}
	}
	return nil
}

func sourceEnvFile(envFilePath string) error {
	LogWarn(fmt.Sprintf("Sourcing environment file from path: %s", envFilePath))

	file, err := os.Open(envFilePath)
	if err != nil {
		return LogError(fmt.Sprintf("Failed to open environment file: %s - %v", envFilePath, err), err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			LogError(fmt.Sprintf("Failed to close environment file: %s - %v", envFilePath, err), err)
		} else {
			LogInfo(fmt.Sprintf("Successfully closed environment file: %s", envFilePath))
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		LogDebug(fmt.Sprintf("Processing line: %s", line))

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return LogError(fmt.Sprintf("Invalid environment variable declaration: %s in file: %s", line, envFilePath), err)
		}

		key := parts[0]
		value := strings.Trim(parts[1], "\"")
		if err := os.Setenv(key, value); err != nil {
			return LogError(fmt.Sprintf("Failed to set environment variable %s: %v - %s", key, err, envFilePath), err)
		}
		LogInfo(fmt.Sprintf("Set environment variable %s=%s", key, value))
	}

	if err := scanner.Err(); err != nil {
		return LogError(fmt.Sprintf("Error reading environment file: %s - %v", envFilePath, err), err)
	}

	LogInfo(fmt.Sprintf("Successfully sourced environment file: %s", envFilePath))
	return nil
}

func setup() (string, func()) {
	workDir := createWorkDir()

	// Set up signal handling to clean up on interruption
	cleanup := func() {
		if err := os.RemoveAll(workDir); err != nil {
			logger.Errorf("Failed to remove work directory: %v", err)
		} else {
			logger.Infof("Cleaned up work directory: %s", workDir)
		}
	}

	envFilePath := filepath.Join(workDir, ".runner_env")
	if err := writeEnvToFile(envFilePath); err != nil {
		logger.Fatalf("Failed to write environment variables to file: %v", err)
	}

	if err := sourceEnvFile(envFilePath); err != nil {
		logger.Fatalf("Failed to source environment file: %s - %v", envFilePath, err)
	}

	// data, err := os.ReadFile(os.Getenv("RUNNER_ENV"))
	// if err != nil {
	//    log.Fatal(err)
	// }
	// os.Stdout.Write(data)

	// Return the setup information and the teardown function
	return workDir, cleanup
}
