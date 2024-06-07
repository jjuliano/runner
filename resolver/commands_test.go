package resolver

import (
	"net/http"
	"path/filepath"
	"sync"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/kdeps/plugins/kdepexec"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

var fs afero.Fs

func setupTestRunResolver() *DependencyResolver {
	fs = afero.NewMemMapFs()

	workDir, teardown := setup()
	defer teardown()

	// Write environment variables to .kdeps_env file
	envFilePath := filepath.Join(workDir, ".kdeps_env")
	if err := writeEnvToFile(envFilePath); err != nil {
		logger.Fatalf("Failed to write environment variables to file: %v", err)
	}

	// Source the .kdeps_env file
	if err := sourceEnvFile(envFilePath); err != nil {
		logger.Fatalf("Failed to source environment file: %v", err)
	}

	logger := log.New(nil)
	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()

	resolver, err := NewGraphResolver(fs, logger, workDir, session)
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	yamlMap := map[string]interface{}{
		"resources": []map[string]interface{}{
			{
				"id":       "homebrew",
				"name":     "Homebrew",
				"desc":     "Homebrew Package Manager",
				"category": "development",
				"requires": []interface{}{},
				"run": []map[string]interface{}{
					{
						"name": "install homebrew",
						"exec": "echo $HELLO",
						"skip": []string{"CMD:bre1w"},
						"env": []map[string]interface{}{
							{
								"name":     "HELLO",
								"kdepexec": "echo 'hello world'",
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
	err := resolver.ProcessNodeSteps(steps, "sampleType", "sampleResNode", client, &KdepsLogs{})

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
			err := ProcessSingleNodeRule(tc.element, httpClient, &KdepsLogs{})

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

	err := ProcessSingleNodeRule(expectedPrefix, client, &KdepsLogs{})
	expectedError := "expected environment variable 'HELLO' does not exist"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got: %v", expectedError, err)
	}
}

func TestProcessElement_Map(t *testing.T) {
	client := &http.Client{}
	expectedExpectations := []interface{}{"unfound value"}

	err := ProcessSingleNodeRule(map[interface{}]interface{}{"expect": expectedExpectations}, client, &KdepsLogs{})
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
			resolver.ProcessNodeSkipRules(tc.step, "test_node", skipResults, mu, client, &KdepsLogs{})

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
	logs := KdepsLogs{}

	// Create a sample log entry
	entry := StepLog{
		name:    "test_step",
		message: "This is a test message",
		id:      "test_resource",
		command: "echo hello world",
	}

	// Add the log entry to the KdepsLogs
	logs.Add(entry)

	// Check if the length of entries slice is 1
	if len(logs.StepLogs()) != 1 {
		t.Errorf("Expected KdepsLogs.Entries to have length 1, got %d", len(logs.StepLogs()))
	}

	logs.Close()

	// Check if the first entry in the entries slice is equal to the sample log entry
	if logs.entries[0] != entry {
		t.Errorf("Expected KdepsLogs.Entries[0] to be equal to entry, but got %+v", logs.entries[0])
	}

	// Check if the log messages are correctly retrieved as a single string
	messages := logs.GetAllMessageString()
	expectedMessage := "This is a test message"
	if messages != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, messages)
	}

}
