package resolver

import (
	"testing"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
)

var testFilePaths = []string{"/test/file1.yaml", "/test/file2.yaml"}

func TestLoadResourceEntries(t *testing.T) {
	fs := afero.NewMemMapFs()
	logger := log.New(nil)
	dr, err := NewDependencyResolver(fs, logger, "")
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	yamlData1 := `
resources:
  - resource: "testres1"
    name: "Test Resource 1"
    sdesc: "A test resource 1"
    ldesc: "A longer description 1"
    category: "test"
    requires:
      - "dep1"
      - "dep2"
`
	yamlData2 := `
resources:
  - resource: "testres2"
    name: "Test Resource 2"
    sdesc: "A test resource 2"
    ldesc: "A longer description 2"
    category: "test"
    requires:
      - "dep3"
      - "dep4"
`
	afero.WriteFile(fs, testFilePaths[0], []byte(yamlData1), 0644)
	afero.WriteFile(fs, testFilePaths[1], []byte(yamlData2), 0644)

	for _, filePath := range testFilePaths {
		dr.LoadResourceEntries(filePath)
	}

	if len(dr.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(dr.Resources))
	}
	if dr.Resources[0].Resource != "testres1" || dr.Resources[1].Resource != "testres2" {
		t.Errorf("Expected resources 'testres1' and 'testres2', got '%s' and '%s'", dr.Resources[0].Resource, dr.Resources[1].Resource)
	}
}

func TestLoadResourceEntries_CircularDependency(t *testing.T) {
	fs := afero.NewMemMapFs()
	logger := log.New(nil)
	dr, err := NewDependencyResolver(fs, logger, "")
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	yamlData := `
resources:
  - resource: "a"
    name: "Resource A"
    sdesc: "Short description A"
    ldesc: "Long description A"
    category: "test"
    requires:
      - "c"
  - resource: "b"
    name: "Resource B"
    sdesc: "Short description B"
    ldesc: "Long description B"
    category: "test"
    requires:
      - "a"
  - resource: "c"
    name: "Resource C"
    sdesc: "Short description C"
    ldesc: "Long description C"
    category: "test"
    requires:
      - "b"
`
	afero.WriteFile(fs, testFilePaths[0], []byte(yamlData), 0644)
	dr.LoadResourceEntries(testFilePaths[0])

	if len(dr.Resources) != 3 {
		t.Errorf("Expected 3 resources, got %d", len(dr.Resources))
	}

	expectedResources := []string{"a", "b", "c"}
	for i, res := range expectedResources {
		if dr.Resources[i].Resource != res {
			t.Errorf("Expected resource %s, got %s", res, dr.Resources[i].Resource)
		}
	}
}
