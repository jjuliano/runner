// main_test.go
package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"kdeps/resolver"

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

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs(args)

	_, err = root.ExecuteC()
	return buf.String(), err
}

func initTestConfig() {
	viper.SetConfigName("test_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		viper.Set("package_files", []string{"./test_packages.yaml"})
	}
}

func setupTestResolver() *resolver.DependencyResolver {
	fs := afero.NewMemMapFs()
	resolver := resolver.NewDependencyResolver(fs)
	yamlData := `
packages:
  - package: "pkg1"
    name: "Package 1"
    sdesc: "Short description 1"
    ldesc: "Long description 1"
    category: "cat1"
    requires: ["pkg2"]
  - package: "pkg2"
    name: "Package 2"
    sdesc: "Short description 2"
    ldesc: "Long description 2"
    category: "cat2"
    requires: ["pkg3"]
  - package: "pkg3"
    name: "Package 3"
    sdesc: "Short description 3"
    ldesc: "Long description 3"
    category: "cat3"
    requires: []
`
	afero.WriteFile(fs, "./test_packages.yaml", []byte(yamlData), 0644)
	resolver.LoadPackageEntries("./test_packages.yaml")
	return resolver
}

func TestDependsCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "depends [package_names...]",
		Short: "List dependencies of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleDependsCommand(args)
		},
	})

	args := []string{"depends", "pkg1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "pkg1\npkg1 > pkg2\npkg1 > pkg2 > pkg3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestRDependsCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "rdepends [package_names...]",
		Short: "List reverse dependencies of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleRDependsCommand(args)
		},
	})

	args := []string{"rdepends", "pkg3"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "pkg3\npkg3 > pkg2\npkg3 > pkg2 > pkg1\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestShowCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "show [package_names...]",
		Short: "Show details of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleShowCommand(args)
		},
	})

	args := []string{"show", "pkg1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "Package: pkg1\nName: Package 1\nShort Description: Short description 1\nLong Description: Long description 1\nCategory: cat1\nRequirements: [pkg2]\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestSearchCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "search [package_names...]",
		Short: "Search for the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleSearchCommand(args)
		},
	})

	args := []string{"search", "Package 1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "Package: pkg1\nName: Package 1\nShort Description: Short description 1\nLong Description: Long description 1\nCategory: cat1\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestCategoryCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "category [package_names...]",
		Short: "List categories of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleCategoryCommand(args)
		},
	})

	args := []string{"category", "cat3"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "pkg3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestTreeCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree [package_names...]",
		Short: "Show dependency tree of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleTreeCommand(args)
		},
	})

	args := []string{"tree", "pkg1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "pkg1 > pkg2 > pkg3\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}

func TestTreeListCommand(t *testing.T) {
	initTestConfig()
	resolver := setupTestResolver()
	rootCmd := &cobra.Command{Use: "kdeps"}
	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree-list [package_names...]",
		Short: "Show dependency tree list of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleTreeListCommand(args)
		},
	})

	args := []string{"tree-list", "pkg1"}
	rootCmd.SetArgs(args)

	output := captureOutput(func() {
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("Failed to execute command: %v", err)
		}
	})

	expectedOutput := "pkg3\npkg2\npkg1\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedOutput, output)
	}
}
