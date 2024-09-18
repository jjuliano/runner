package resolver

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
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

func createWorkDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "runner_workdir")
	if err != nil {
		return "", err
	}
	return tmpDir, nil
}

func writeEnvToFile(envFilePath string) error {
	envFile, err := os.Create(envFilePath)
	if err != nil {
		return err
	}
	defer envFile.Close()

	os.Setenv("RUNNER_ENV", envFilePath)

	for _, env := range os.Environ() {
		// Split the environment variable into key and value
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]
		value := parts[1]

		// Quote the value if it contains special characters or spaces
		if strings.ContainsAny(value, " \t\n\r\"'") {
			value = strconv.Quote(value)
		}

		// Write the environment variable to the file
		if _, err := envFile.WriteString(fmt.Sprintf("%s=%s\n", key, value)); err != nil {
			return err
		}
	}
	return nil
}

func sourceEnvFile(envFilePath string) error {
	cmd := exec.Command("sh", "-c", "set -a && source "+envFilePath+" && set +a")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func setup() (string, func()) {
	workDir, err := createWorkDir()
	if err != nil {
		logger.Fatalf("Failed to create work directory: %v", err)
	}

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
		logger.Fatalf("Failed to source environment file: %v", err)
	}

	// data, err := os.ReadFile(os.Getenv("RUNNER_ENV"))
	// if err != nil {
	//	log.Fatal(err)
	// }
	// os.Stdout.Write(data)

	// Return the setup information and the teardown function
	return workDir, cleanup
}
