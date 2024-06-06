// main_test.go
package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"kdeps/resolver"

	"github.com/charmbracelet/log"
	"github.com/kdeps/plugins/kdepexec"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
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

func initTestConfig() {
	viper.SetConfigName("test_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		viper.Set("resource_files", []string{"./test_resources.yaml"})
	}
}

func setupTestResolver() *resolver.DependencyResolver {
	fs := afero.NewMemMapFs()
	logger := log.New(nil)
	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()

	resolver, err := resolver.NewDependencyResolver(fs, logger, "", session)
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	yamlData := `
resources:
  - resource: "res1"
    name: "Resource 1"
    sdesc: "Short description 1"
    ldesc: "Long description 1"
    category: "cat1"
    requires: ["res2"]
  - resource: "res2"
    name: "Resource 2"
    sdesc: "Short description 2"
    ldesc: "Long description 2"
    category: "cat2"
    requires: ["res3"]
  - resource: "res3"
    name: "Resource 3"
    sdesc: "Short description 3"
    ldesc: "Long description 3"
    category: "cat3"
    requires: []
`
	afero.WriteFile(fs, "./test_resources.yaml", []byte(yamlData), 0644)
	resolver.LoadResourceEntries("./test_resources.yaml")
	return resolver
}

func TestDependsCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "depends [resource_names...]",
		Short: "List dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleDependsCommand(args)
		},
	})

	args := []string{"depends", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to kdepexecute command: %v", err)
		}
	})

	expectedOutput := "res1\nres1 -> res2\nres1 -> res2 -> res3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestRDependsCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "rdepends [resource_names...]",
		Short: "List reverse dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleRDependsCommand(args)
		},
	})

	args := []string{"rdepends", "res3"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to kdepexecute command: %v", err)
		}
	})

	expectedOutput := "res3\nres3 -> res2\nres3 -> res2 -> res1\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestShowCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "show [resource_names...]",
		Short: "Show details of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleShowCommand(args)
		},
	})

	args := []string{"show", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to kdepexecute command: %v", err)
		}
	})

	expectedOutput := "📦 Resource: res1\n📛 Name: Resource 1\n📝 Short Description: Short description 1\n📖 Long Description: Long description 1\n🏷️  Category: cat1\n🔗 Requirements: [res2]\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestSearchCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "search [resource_names...]",
		Short: "Search for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleSearchCommand(args)
		},
	})

	args := []string{"search", "Resource 1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to kdepexecute command: %v", err)
		}
	})

	expectedOutput := "📦 Resource: res1\n📛 Name: Resource 1\n📝 Short Description: Short description 1\n📖 Long Description: Long description 1\n🏷️  Category: cat1\n🔗 Requirements: [res2]\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestCategoryCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "category [resource_names...]",
		Short: "List categories of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleCategoryCommand(args)
		},
	})

	args := []string{"category", "cat3"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to kdepexecute command: %v", err)
		}
	})

	expectedOutput := "res3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestTreeCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree [resource_names...]",
		Short: "Show dependency tree of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleTreeCommand(args)
		},
	})

	args := []string{"tree", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to kdepexecute command: %v", err)
		}
	})

	expectedOutput := "res1 <- res2 <- res3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestTreeListCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree-list [resource_names...]",
		Short: "Show dependency tree list of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleTreeListCommand(args)
		},
	})

	args := []string{"tree-list", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to kdepexecute command: %v", err)
		}
	})

	expectedOutput := "res3\nres2\nres1\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestDependsCommand_CircularDependency(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "depends [resource_names...]",
		Short: "List dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleDependsCommand(args)
		},
	})

	args := []string{"depends", "res1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to kdepexecute command: %v", err)
		}
	})

	expectedOutput := "res1\nres1 -> res2\nres1 -> res2 -> res3"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}
