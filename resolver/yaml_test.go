package resolver

import (
	"testing"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

var testFilePath = "/test/packages.yaml"

func TestLoadPackageEntries(t *testing.T) {
	fs := afero.NewMemMapFs()
	dr := NewDependencyResolver(fs)

	yamlData := `
packages:
  - package: "testpkg"
    name: "Test Package"
    sdesc: "A test package"
    ldesc: "A longer description"
    category: "test"
    requires:
      - "dep1"
      - "dep2"
`
	afero.WriteFile(fs, testFilePath, []byte(yamlData), 0644)

	dr.LoadPackageEntries(testFilePath)

	if len(dr.Packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(dr.Packages))
	}
	if dr.Packages[0].Package != "testpkg" {
		t.Errorf("Expected package 'testpkg', got '%s'", dr.Packages[0].Package)
	}
}

func TestSavePackageEntries(t *testing.T) {
	fs := afero.NewMemMapFs()
	dr := NewDependencyResolver(fs)

	dr.Packages = []PackageEntry{
		{Package: "testpkg", Name: "Test Package", Sdesc: "A test package", Ldesc: "A longer description", Category: "test", Requires: []string{"dep1", "dep2"}},
	}

	dr.SavePackageEntries(testFilePath)

	content, err := afero.ReadFile(fs, testFilePath)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	var data struct {
		Packages []PackageEntry `yaml:"packages"`
	}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		t.Errorf("Error unmarshalling YAML: %v", err)
	}

	if len(data.Packages) != 1 {
		t.Errorf("Expected 1 package, got %d", len(data.Packages))
	}
	if data.Packages[0].Package != "testpkg" {
		t.Errorf("Expected package 'testpkg', got '%s'", data.Packages[0].Package)
	}
}
