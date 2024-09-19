package check

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jjuliano/runner/pkg/expect/process"
	"github.com/jjuliano/runner/pkg/kdepexec"
)

// addDefaultProtocol ensures the URL has a protocol. If missing, it adds "http://"
func addDefaultProtocol(url string) string {
	if !strings.Contains(url, "://") {
		return "http://" + url
	}
	return url
}

// CheckExpectations verifies if the output or exit code matches the expectations.
func CheckExpectations(output string, exitCode int, expectations []string, client *http.Client) error {
	for _, exp := range expectations {
		isNegation := strings.HasPrefix(exp, "!")
		expectation := strings.TrimPrefix(exp, "!")
		expectation = process.ReplaceVars(expectation)

		// Check if the expectation is a command to execute
		if strings.HasPrefix(expectation, "CMD:") {
			cmd := strings.TrimPrefix(expectation, "CMD:")
			path, err := kdepexec.Which(cmd)
			if isNegation {
				if err == nil {
					return fmt.Errorf("unexpected executable path '%s' exists", path)
				}
			} else {
				if err != nil {
					return fmt.Errorf("expected executable path '%s' does not exist", cmd)
				}
			}
			continue
		}

		// Check if the expectation is a command to execute and verify its exit status
		if strings.HasPrefix(expectation, "EXEC:") {
			cmdStr := strings.TrimPrefix(expectation, "EXEC:")
			cmdParts := strings.Fields(cmdStr)
			if len(cmdParts) == 0 {
				return fmt.Errorf("invalid EXEC command")
			}

			cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
			output, err := cmd.CombinedOutput()
			if isNegation {
				if err == nil {
					return fmt.Errorf("unexpected command '%s' ran successfully: %s", cmdStr, output)
				}
			} else {
				if err != nil {
					return fmt.Errorf("command '%s' failed: %v, output: %s", cmdStr, err, output)
				}
			}
			continue
		}

		// Check if the expectation is an exit code (number)
		if expectNum, err := strconv.Atoi(expectation); err == nil {
			if isNegation {
				if exitCode == expectNum {
					return fmt.Errorf("unexpected exit status '%d'", exitCode)
				}
			} else {
				if exitCode != expectNum {
					return fmt.Errorf("expected exit status '%d' but got '%d'", expectNum, exitCode)
				}
			}
			continue
		}

		// Check if the expectation is an environment variable
		if strings.HasPrefix(expectation, "ENV:") {
			envVar := strings.TrimPrefix(expectation, "ENV:")
			_, exists := os.LookupEnv(envVar)
			if isNegation {
				if exists {
					return fmt.Errorf("unexpected environment variable '%s' exists", envVar)
				}
			} else {
				if !exists {
					return fmt.Errorf("expected environment variable '%s' does not exist", envVar)
				}
			}
			continue
		}

		// Check if the expectation is a file
		if strings.HasPrefix(expectation, "FILE:") {
			filePath := strings.TrimPrefix(expectation, "FILE:")
			if isNegation {
				if _, err := os.Stat(filePath); err == nil {
					return fmt.Errorf("unexpected file '%s' exists", filePath)
				}
			} else {
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					return fmt.Errorf("expected file '%s' does not exist", filePath)
				}
			}
			continue
		}

		// Check if the expectation is a directory
		if strings.HasPrefix(expectation, "DIR:") {
			dirPath := strings.TrimPrefix(expectation, "DIR:")
			if isNegation {
				if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
					return fmt.Errorf("unexpected directory '%s' exists", dirPath)
				}
			} else {
				if info, err := os.Stat(dirPath); os.IsNotExist(err) || !info.IsDir() {
					return fmt.Errorf("expected directory '%s' does not exist", dirPath)
				}
			}
			continue
		}

		// Check if the expectation is a URL
		if strings.HasPrefix(expectation, "URL:") {
			url := strings.TrimPrefix(expectation, "URL:")
			url = addDefaultProtocol(url) // Ensure the URL has a protocol
			resp, err := client.Head(url)
			if isNegation {
				if err == nil && resp.StatusCode == http.StatusOK {
					return fmt.Errorf("unexpected URL '%s' is accessible", url)
				}
			} else {
				if err != nil || resp.StatusCode != http.StatusOK {
					return fmt.Errorf("expected URL '%s' is not accessible", url)
				}
			}
			continue
		}

		// Default string expectation check
		if isNegation {
			if strings.Contains(strings.ToLower(output), strings.ToLower(expectation)) {
				return fmt.Errorf("unexpected output: found '%s'", expectation)
			}
		} else {
			if !strings.Contains(strings.ToLower(output), strings.ToLower(expectation)) {
				return fmt.Errorf("expected '%s' not found in output", expectation)
			}
		}
	}
	return nil
}
