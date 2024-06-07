package resolver

import (
	"testing"

	"github.com/charmbracelet/log"
	"github.com/kdeps/plugins/kdepexec"
	"github.com/spf13/afero"
)

var testFilePaths = []string{"/test/file1.yaml", "/test/file2.yaml"}

func TestLoadResourceEntries(t *testing.T) {
	fs := afero.NewMemMapFs()
	logger := log.New(nil)
	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()
	dr, err := NewGraphResolver(fs, logger, "", session)
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	yamlData1 := `
resources:
  - id: "testres1"
    name: "Test Id 1"
    desc: "A longer description 1"
    category: "test"
    requires:
      - "dep1"
      - "dep2"
`
	yamlData2 := `
resources:
  - id: "testres2"
    name: "Test Id 2"
    desc: "A longer description 2"
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
	if dr.Resources[0].Id != "testres1" || dr.Resources[1].Id != "testres2" {
		t.Errorf("Expected resources 'testres1' and 'testres2', got '%s' and '%s'", dr.Resources[0].Id, dr.Resources[1].Id)
	}
}

func TestLoadResourceEntries_CircularDependency(t *testing.T) {
	fs := afero.NewMemMapFs()
	logger := log.New(nil)
	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()
	dr, err := NewGraphResolver(fs, logger, "", session)
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	yamlData := `
resources:
  - id: "a"
    name: "Id A"
    desc: "Long description A"
    category: "test"
    requires:
      - "c"
  - id: "b"
    name: "Id B"
    desc: "Long description B"
    category: "test"
    requires:
      - "a"
  - id: "c"
    name: "Id C"
    desc: "Long description C"
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
		if dr.Resources[i].Id != res {
			t.Errorf("Expected resource %s, got %s", res, dr.Resources[i].Id)
		}
	}
}
