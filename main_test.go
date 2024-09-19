// main_test.go
package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjuliano/runner/pkg/resolver"

	"github.com/charmbracelet/log"
	"github.com/jjuliano/runner/pkg/kdepexec"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// Helper function to capture command output
func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func initTestConfig(t *testing.T) (afero.Fs, string, string) {
	fs := afero.NewOsFs()
	tempDir := filepath.Join(os.TempDir(), "runner_test")
	localFile := filepath.Join(tempDir, "test_resources.yaml")
	configFile := filepath.Join(tempDir, "test_config.yaml")

	// Create the temp directory
	err := fs.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create the test_resources.yaml file
	file, err := fs.Create(localFile)
	if err != nil {
		t.Fatalf("failed to create local file: %v", err)
	}
	file.Close()

	// Create the test_config.yaml file
	file, err = fs.Create(configFile)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	file.Close()

	// Defer the deletion of the temp directory
	t.Cleanup(func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("failed to remove temp dir: %v", err)
		}
	})

	return fs, configFile, localFile
}

func setupTestResolver(fs afero.Fs, configFile string, localFile string) *resolver.DependencyResolver {
	logger := log.New(nil)

	localYAMLContent := []byte(`
resources:
  - id: "res1"
    name: "Id 1"
    desc: "Long description 1"
    category: "cat1"
    requires: ["res2"]
  - id: "res2"
    name: "Id 2"
    desc: "Long description 2"
    category: "cat2"
    requires: ["res3"]
  - id: "res3"
    name: "Id 3"
    desc: "Long description 3"
    category: "cat3"
    requires: []
`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(localYAMLContent))
	}))

	runnerConfigContent := []byte(`
workflows:
  - ` + server.URL + `
  - ` + localFile + `
`)

	afero.WriteFile(fs, configFile, []byte(runnerConfigContent), 0644)
	afero.WriteFile(fs, localFile, []byte(localYAMLContent), 0644)
	defer server.Close()

	viper.SetConfigName(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		resolver.PrintError("Error reading config file", err)
		os.Exit(1)
	}

	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()

	dependencyResolver, err := resolver.NewGraphResolver(fs, logger, "", session)
	if err != nil {
		log.Fatalf("Failed to create dependency dependencyResolver: %v", err)
	}

	resourceFiles := viper.GetStringSlice("workflows")
	for _, file := range resourceFiles {
		if err := dependencyResolver.LoadResourceEntries(file); err != nil {
			resolver.PrintError("Error loading resource entries", err)
			os.Exit(1)
		}
	}

	return dependencyResolver
}

func TestDependsCommand(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"depends", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "res1\nres1 -> res2\nres1 -> res2 -> res3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestRDependsCommand(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"rdepends", "res3"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "res3\nres3 -> res2\nres3 -> res2 -> res1\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestShowCommand(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"show", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "📦 Id: res1\n📛 Name: Id 1\n📝 Description: Long description 1\n🏷️  Category: cat1\n🔗 Requirements: [res2]\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestSearchCommand(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"search", "Id 1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "📦 Id: res1\n📛 Name: Id 1\n📝 Description: Long description 1\n🏷️  Category: cat1\n🔗 Requirements: [res2]\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestCategoryCommand(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"category", "cat3"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "res3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestTreeCommand(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"tree", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "res1 <- res2 <- res3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestTreeListCommand(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"tree-list", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "res3\nres2\nres1\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestDependsCommand_CircularDependency(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"depends", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "res1\nres1 -> res2\nres1 -> res2 -> res3"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestIndexCommand(t *testing.T) {
	resolver := setupTestResolver(initTestConfig(t))
	rootCmd := createRootCmd(resolver)

	args := []string{"index"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput :=
		`📦 Id: res1
📛 Name: Id 1
📝 Description: Long description 1
🏷️  Category: cat1
🔗 Requirements: [res2]
---
📦 Id: res2
📛 Name: Id 2
📝 Description: Long description 2
🏷️  Category: cat2
🔗 Requirements: [res3]
---
📦 Id: res3
📛 Name: Id 3
📝 Description: Long description 3
🏷️  Category: cat3
🔗 Requirements: []
---
📦 Id: res1
📛 Name: Id 1
📝 Description: Long description 1
🏷️  Category: cat1
🔗 Requirements: [res2]
---
📦 Id: res2
📛 Name: Id 2
📝 Description: Long description 2
🏷️  Category: cat2
🔗 Requirements: [res3]
---
📦 Id: res3
📛 Name: Id 3
📝 Description: Long description 3
🏷️  Category: cat3
🔗 Requirements: []
---
`
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}
