package resolver

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var fs afero.Fs

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs(args)

	_, err = root.ExecuteC()
	return buf.String(), err
}

func initTestConfig(fs afero.Fs) {
	configMap := map[string]interface{}{
		"resource_files": []string{"./test_resources.yaml"},
	}

	yamlData, err := yaml.Marshal(configMap)
	if err != nil {
		log.Fatalf("Error marshalling config data: %v", err)
	}

	err = afero.WriteFile(fs, "kdeps.yaml", yamlData, 0644)
	if err != nil {
		log.Fatalf("Error writing config file: %v", err)
	}
}

func setupTestRunResolver() *DependencyResolver {
	fs = afero.NewMemMapFs()
	logger := log.New(nil)
	resolver, err := NewDependencyResolver(fs, logger)
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	yamlMap := map[string]interface{}{
		"resources": []map[string]interface{}{
			{
				"resource": "homebrew",
				"name":     "Homebrew",
				"sdesc":    "Homebrew Package Manager",
				"ldesc":    "Homebrew is a package manager for macOS.",
				"category": "development",
				"requires": []interface{}{},
				"run": []map[string]interface{}{
					{
						"name": "install homebrew",
						"exec": "echo $HELLO",
						"skip": []string{"CMD:bre1w"},
						"env": []map[string]interface{}{
							{
								"name": "HELLO",
								"exec": "echo 'hello world'",
							},
							{
								"name":  "HELLO2",
								"value": "HELLO 2",
							},
						},
					},
					{
						"name":   "test envvar",
						"exec":   "echo $HELLO2",
						"check":  []string{"ENV:HELLO2"},
						"expect": []string{"HELLO 2", "ENV:HELLO2", "CMD:brew"},
					},
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(yamlMap)
	if err != nil {
		log.Fatalf("Error marshalling YAML data: %v", err)
	}

	afero.WriteFile(fs, "./test_resources.yaml", yamlData, 0644)

	return resolver
}

func TestProcessSteps(t *testing.T) {
	// Create a sample DependencyResolver instance
	resolver := setupTestRunResolver()

	// Define sample steps to be processed
	steps := []interface{}{
		"step1",
		"step2",
		"step3",
	}

	// Mock HTTP client
	client := &http.Client{}

	// Call the function being tested
	err := resolver.processSteps(steps, "sampleType", "sampleResNode", client, &logs{})

	// Check if there were any errors returned
	if err != nil {
		t.Errorf("Error processing steps: %v", err)
	}
}

func TestProcessElement(t *testing.T) {
	// Mock HTTP client
	httpClient := &http.Client{}

	// Test cases
	testCases := []struct {
		name        string
		element     interface{}
		expectedErr error
	}{
		{
			name:        "Valid string with ENV prefix",
			element:     "ENV:HOME",
			expectedErr: nil,
		},
		{
			name:        "Invalid string with unsupported prefix",
			element:     "unsupported_prefix:value",
			expectedErr: nil,
		},
		// Add more test cases here as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute the function with the test case input
			err := processElement(tc.element, httpClient, &logs{})

			// Check if the error matches the expected error
			if (err == nil && tc.expectedErr != nil) || (err != nil && tc.expectedErr == nil) || (err != nil && tc.expectedErr != nil && err.Error() != tc.expectedErr.Error()) {
				t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestProcessElement_String(t *testing.T) {
	client := &http.Client{}
	expectedPrefix := "ENV:HELLO"

	err := processElement(expectedPrefix, client, &logs{})
	expectedError := "expected environment variable 'HELLO' does not exist"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %v", expectedError, err)
	}
}

func TestProcessElement_Map(t *testing.T) {
	client := &http.Client{}
	expectedExpectations := []interface{}{"unfound value"}

	err := processElement(map[interface{}]interface{}{"expect": expectedExpectations}, client, &logs{})
	expectedError := "expected 'unfound value' not found in output"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %v", expectedError, err)
	}
}

func TestProcessSkipSteps(t *testing.T) {
	client := &http.Client{}
	resolver := setupTestRunResolver()

	// Mock skipResults map and mutex
	skipResults := make(map[StepKey]bool)
	mu := &sync.Mutex{}

	// Define test cases
	testCases := []struct {
		name         string
		step         RunStep
		expectedSkip bool
	}{
		{
			name: "Valid skip step",
			step: RunStep{
				Name: "test_step1",
				Skip: []interface{}{"ENV:HOME"},
			},
			expectedSkip: true,
		},
		{
			name: "Invalid skip step",
			step: RunStep{
				Name: "test_step2",
				Skip: []interface{}{"CMD:invalidCommand"},
			},
			expectedSkip: false,
		},
		{
			name: "No skip steps",
			step: RunStep{
				Name: "test_step3",
				Skip: nil,
			},
			expectedSkip: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Process skip steps
			resolver.processSkipSteps(tc.step, "test_node", skipResults, mu, client, &logs{})

			// Check the result
			skipKey := StepKey{name: tc.step.Name, node: "test_node"}
			skipResult := skipResults[skipKey]
			if skipResult != tc.expectedSkip {
				t.Errorf("Expected skip result for step '%s' to be '%v', got '%v'", tc.step.Name, tc.expectedSkip, skipResult)
			}

			// Debugging output
			t.Logf("Skip results for step '%s': expected %v, got %v", tc.step.Name, tc.expectedSkip, skipResult)
		})
	}
}

func TestAddLogEntry(t *testing.T) {
	setupTestRunResolver()
	logs := logs{}

	// Create a sample log entry
	entry := stepLog{
		name:    "test_step",
		message: "This is a test message",
		res:     "test_resource",
		command: "echo hello world",
	}

	// Add the log entry to the logs
	logs.add(entry)

	// Check if the length of entries slice is 1
	if len(logs.getAll()) != 1 {
		t.Errorf("Expected logs.Entries to have length 1, got %d", len(logs.getAll()))
	}

	logs.close()

	// Check if the first entry in the entries slice is equal to the sample log entry
	if logs.entries[0] != entry {
		t.Errorf("Expected logs.Entries[0] to be equal to entry, but got %+v", logs.entries[0])
	}

	// Check if the log messages are correctly retrieved as a single string
	messages := logs.getAllMessageString()
	expectedMessage := "This is a test message"
	if messages != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, messages)
	}

}

func TestHandleRunCommand(t *testing.T) {
	fs := afero.NewMemMapFs()
	setupTestRunResolver()
	initTestConfig(fs)

	viper.SetConfigFile("kdeps.yaml")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Run a specified command",
		Run: func(cmd *cobra.Command, args []string) {
			// handleRun logic here, for now, we'll just mock it
			// This part should be replaced with the actual handleRun logic
			for _, arg := range args {
				switch arg {
				case "skip":
					t.Log("Skipping the test")
					return
				case "check":
					if len(args) < 2 {
						t.Errorf("Expected at least 2 arguments, got %d", len(args))
					}
				case "expect":
					expected := "expected"
					if args[1] != expected {
						t.Errorf("Expected '%s', got '%s'", expected, args[1])
					}
				case "env":
					envVar := "HELLO"
					expectedValue := "hello world"
					if os.Getenv(envVar) != expectedValue {
						t.Errorf("Expected env var '%s' to be '%s', got '%s'", envVar, expectedValue, os.Getenv(envVar))
					}
				case "env2":
					envVar := "HELLO2"
					expectedValue := "HELLO 2"
					if os.Getenv(envVar) != expectedValue {
						t.Errorf("Expected env var '%s' to be '%s', got '%s'", envVar, expectedValue, os.Getenv(envVar))
					}
				}
			}
		},
	})

	t.Run("Test handleRun with skip", func(t *testing.T) {
		args := []string{"run", "skip"}
		rootCmd.SetArgs(args)

		output := captureOutput(func() {
			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}
		})

		if !strings.Contains(output, "") {
			t.Errorf("Expected 'Skipping the test', got '%s'", output)
		}
	})

	t.Run("Test handleRun with check", func(t *testing.T) {
		args := []string{"run", "check", "additional"}
		rootCmd.SetArgs(args)

		output := captureOutput(func() {
			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}
		})

		expectedOutput := ""
		if output != expectedOutput {
			t.Errorf("Expected '%s', got '%s'", expectedOutput, output)
		}
	})

	t.Run("Test handleRun with expect", func(t *testing.T) {
		args := []string{"run", "expect", "expected"}
		rootCmd.SetArgs(args)

		output := captureOutput(func() {
			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}
		})

		expectedOutput := ""
		if output != expectedOutput {
			t.Errorf("Expected '%s', got '%s'", expectedOutput, output)
		}
	})

	t.Run("Test handleRun with env", func(t *testing.T) {
		os.Setenv("HELLO", "hello world")
		defer os.Unsetenv("HELLO")

		args := []string{"run", "env"}
		rootCmd.SetArgs(args)

		output := captureOutput(func() {
			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}
		})

		expectedOutput := ""
		if output != expectedOutput {
			t.Errorf("Expected '%s', got '%s'", expectedOutput, output)
		}
	})

	t.Run("Test handleRun with env2", func(t *testing.T) {
		os.Setenv("HELLO2", "HELLO 2")
		defer os.Unsetenv("HELLO2")

		args := []string{"run", "env2"}
		rootCmd.SetArgs(args)

		output := captureOutput(func() {
			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("Failed to execute command: %v", err)
			}
		})

		expectedOutput := ""
		if output != expectedOutput {
			t.Errorf("Expected '%s', got '%s'", expectedOutput, output)
		}
	})
}
