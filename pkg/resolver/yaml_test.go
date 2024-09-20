package resolver

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/jjuliano/runner/pkg/kdepexec"
	"github.com/spf13/afero"
)

func TestLoadResourceEntries(t *testing.T) {
	logger := log.New(nil)
	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()

	dr, err := NewGraphResolver(afero.NewOsFs(), logger, "", session)
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

	tmpFile1, err := ioutil.TempFile("", "resource1-*.yaml")
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile1.Name())

	tmpFile2, err := ioutil.TempFile("", "resource2-*.yaml")
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile2.Name())

	if _, err := tmpFile1.Write([]byte(yamlData1)); err != nil {
		log.Fatalf("Failed to write to temp file: %v", err)
	}
	if _, err := tmpFile2.Write([]byte(yamlData2)); err != nil {
		log.Fatalf("Failed to write to temp file: %v", err)
	}

	testFilePaths := []string{tmpFile1.Name(), tmpFile2.Name()}

	for _, filePath := range testFilePaths {
		dr.LoadResourceEntries(filePath)
	}

	tmpFile1.Close()
	tmpFile2.Close()

	if len(dr.Resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(dr.Resources))
	}
	if dr.Resources[0].Id != "testres1" || dr.Resources[1].Id != "testres2" {
		t.Errorf("Expected resources 'testres1' and 'testres2', got '%s' and '%s'", dr.Resources[0].Id, dr.Resources[1].Id)
	}
}

func TestLoadResourceEntries_CircularDependency(t *testing.T) {
	logger := log.New(nil)
	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()

	dr, err := NewGraphResolver(afero.NewOsFs(), logger, "", session)
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

	tmpFile, err := ioutil.TempFile("", "test1-*.yaml")
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		log.Fatalf("Failed to write to temp file: %v", err)
	}

	testFilePaths := []string{tmpFile.Name()}

	for _, filePath := range testFilePaths {
		dr.LoadResourceEntries(filePath)
	}

	afero.WriteFile(fs, testFilePaths[0], []byte(yamlData), 0644)
	dr.LoadResourceEntries(testFilePaths[0])

	tmpFile.Close()

	expectedResources := []string{"a", "b", "c"}
	for i, res := range expectedResources {
		if dr.Resources[i].Id != res {
			t.Errorf("Expected resource %s, got %s", res, dr.Resources[i].Id)
		}
	}
}
