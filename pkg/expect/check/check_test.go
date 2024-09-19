package check

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCheckExpectations(t *testing.T) {
	client := &http.Client{}

	t.Run("Test Exit Code Expectations", func(t *testing.T) {
		expectations := []string{"0"}
		output := ""
		exitCode := 0

		err := CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectations = []string{"1"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("Test Output String Expectations", func(t *testing.T) {
		expectations := []string{"hello"}
		output := "hello world"
		exitCode := 0

		err := CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectations = []string{"!hello"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("Test Environment Variable Expectations", func(t *testing.T) {
		os.Setenv("TEST_VAR", "value")
		defer os.Unsetenv("TEST_VAR")

		expectations := []string{"ENV:TEST_VAR"}
		output := ""
		exitCode := 0

		err := CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectations = []string{"ENV:NON_EXISTENT_VAR"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("Test File Expectations", func(t *testing.T) {
		file, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(file.Name())

		expectations := []string{"FILE:" + file.Name()}
		output := ""
		exitCode := 0

		err = CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectations = []string{"FILE:/non/existent/file"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("Test Directory Expectations", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "testdir")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(dir)

		expectations := []string{"DIR:" + dir}
		output := ""
		exitCode := 0

		err = CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectations = []string{"DIR:/non/existent/dir"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("Test URL Expectations", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		expectations := []string{"URL:" + ts.URL}
		output := ""
		exitCode := 0

		err := CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectations = []string{"URL:http://non.existent.url"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("Test CMD Expectations", func(t *testing.T) {
		// This should pass if `ls` exists on the system
		expectations := []string{"CMD:ls"}
		output := ""
		exitCode := 0

		err := CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// This should fail because `nonexistentcmd` should not exist
		expectations = []string{"CMD:nonexistentcmd"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("Test EXEC Expectations", func(t *testing.T) {
		// This should pass if `echo` command exists and runs successfully
		expectations := []string{"EXEC:echo hello"}
		output := ""
		exitCode := 0

		err := CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// This should fail because `nonexistentcmd` should not exist
		expectations = []string{"EXEC:nonexistentcmd"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}

		// This should pass as negation of a non-existent command
		expectations = []string{"!EXEC:nonexistentcmd"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// This should fail as negation of a valid command
		expectations = []string{"!EXEC:echo hello"}
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("Test ENVVAR Expansion", func(t *testing.T) {
		os.Setenv("EXPAND_TEST", "expanded")
		defer os.Unsetenv("EXPAND_TEST")

		// This should result in the string "expanded"
		expectations := []string{"${EXPAND_TEST}"}
		output := "this is an expanded test"
		exitCode := 0

		// Add some debug logging
		fmt.Printf("Expectations: %v\n", expectations)
		fmt.Printf("Output: %v\n", output)
		fmt.Printf("ExitCode: %v\n", exitCode)

		err := CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Ensure that negative expectations are also tested
		expectations = []string{"!${EXPAND_TEST}"}
		output = "this is not an expand_env test"
		err = CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Test with prefix and expansion
		expectations = []string{"TEST: ${EXPAND_TEST}"}
		output = "this is a TEST: expanded test"
		err = CheckExpectations(output, exitCode, expectations, client)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Negative test with prefix and expansion
		expectations = []string{"!TEST: ${EXPAND_TEST}"}
		output = "this is not a TEST: expanded test"
		err = CheckExpectations(output, exitCode, expectations, client)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
}
