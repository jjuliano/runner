package resolver

import (
	"testing"

	"github.com/spf13/afero"
)

var testFilePaths = []string{"/test/file1.yaml", "/test/file2.yaml"}

func TestLoadPackageEntries(t *testing.T) {
	fs := afero.NewMemMapFs()
	dr := NewDependencyResolver(fs)

	yamlData1 := `
packages:
  - package: "testpkg1"
    name: "Test Package 1"
    sdesc: "A test package 1"
    ldesc: "A longer description 1"
    category: "test"
    requires:
      - "dep1"
      - "dep2"
`
	yamlData2 := `
packages:
  - package: "testpkg2"
    name: "Test Package 2"
    sdesc: "A test package 2"
    ldesc: "A longer description 2"
    category: "test"
    requires:
      - "dep3"
      - "dep4"
`
	afero.WriteFile(fs, testFilePaths[0], []byte(yamlData1), 0644)
	afero.WriteFile(fs, testFilePaths[1], []byte(yamlData2), 0644)

	for _, filePath := range testFilePaths {
		dr.LoadPackageEntries(filePath)
	}

	if len(dr.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(dr.Packages))
	}
	if dr.Packages[0].Package != "testpkg1" || dr.Packages[1].Package != "testpkg2" {
		t.Errorf("Expected packages 'testpkg1' and 'testpkg2', got '%s' and '%s'", dr.Packages[0].Package, dr.Packages[1].Package)
	}
}
