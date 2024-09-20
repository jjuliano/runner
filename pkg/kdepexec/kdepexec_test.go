package kdepexec

import (
	"testing"
)

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		cmd      string
		expected []string // Change to a slice to handle multiple expected outputs
		exitCode int
		hasError bool
	}{
		{"echo Hello, World!", []string{"Hello, World!\n"}, 0, false},
		{"invalid_command", []string{
			"sh: invalid_command: command not found\n",
			"sh: 1: invalid_command: not found\n",
		}, 127, true},
		{"exit 2", []string{""}, 2, true},
	}

	for _, test := range tests {
		session, err := NewShellSession()
		if err != nil {
			t.Fatalf("Failed to create shell session: %v", err)
		}
		defer session.Close()

		resultChan := session.ExecuteCommand(test.cmd)
		result := <-resultChan

		// Check if the output matches any of the expected outputs
		outputMatched := false
		for _, expectedOutput := range test.expected {
			if result.Output == expectedOutput {
				outputMatched = true
				break
			}
		}

		if !outputMatched {
			t.Errorf("expected one of %q, got %q", test.expected, result.Output)
		}

		if result.ExitCode != test.exitCode {
			t.Errorf("expected exit code %d, got %d", test.exitCode, result.ExitCode)
		}
		if (result.Err != nil) != test.hasError {
			t.Errorf("expected error %v, got %v", test.hasError, result.Err)
		}
	}
}
