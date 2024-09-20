package kdepexec

import (
	"strings"
	"testing"
)

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		cmd      string
		expected string
		exitCode int
		hasError bool
	}{
		{"echo Hello, World!", "Hello, World!\n", 0, false},
		{"invalid_command", "sh: invalid_command: command not found\n", 127, true}, // Existing test
		{"exit 2", "", 2, true},
	}

	for _, test := range tests {
		session, err := NewShellSession()
		if err != nil {
			t.Fatalf("Failed to create shell session: %v", err)
		}
		defer session.Close()

		resultChan := session.ExecuteCommand(test.cmd)
		result := <-resultChan

		// Handle the specific case for invalid command output
		if test.cmd == "invalid_command" {
			if !(result.Output == test.expected || strings.TrimSuffix(result.Output, "\n") == strings.TrimSuffix(test.expected, "\n")) {
				t.Errorf("expected output %q, got %q", test.expected, result.Output)
			}
		} else {
			if result.Output != test.expected {
				t.Errorf("expected output %q, got %q", test.expected, result.Output)
			}
		}

		if result.ExitCode != test.exitCode {
			t.Errorf("expected exit code %d, got %d", test.exitCode, result.ExitCode)
		}
		if (result.Err != nil) != test.hasError {
			t.Errorf("expected error %v, got %v", test.hasError, result.Err)
		}
	}
}
