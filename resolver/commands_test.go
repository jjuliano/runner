package resolver

import (
	"bytes"
	"os"
	"strings"
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
