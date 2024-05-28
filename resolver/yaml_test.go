package resolver

import (
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

// Sample data for testing
var samplePackages = []PackageEntry{
	{Package: "a", Name: "A", Sdesc: "Package A", Ldesc: "The first package in the alphabetical order", Category: "example", Requires: []string{}},
	{Package: "b", Name: "B", Sdesc: "Package B", Ldesc: "The second package, dependent on A", Category: "example", Requires: []string{"a"}},
	{Package: "c", Name: "C", Sdesc: "Package C", Ldesc: "The third package, dependent on B", Category: "example", Requires: []string{"b"}},
	// Add more entries as needed
}

// Test file path
const testFilePath = "test_packages.yaml"

func setupTestFile() {
	data := struct {
		Packages []PackageEntry `yaml:"packages"`
	}{
		Packages: samplePackages,
	}

	content, _ := yaml.Marshal(data)
	ioutil.WriteFile(testFilePath, content, 0644)
}

func cleanupTestFile() {
	os.Remove(testFilePath)
}

func TestLoadPackageEntries(t *testing.T) {
	setupTestFile()
	defer cleanupTestFile()

	resolver := NewDependencyResolver()
	resolver.LoadPackageEntries(testFilePath)

	if len(resolver.Packages) != len(samplePackages) {
		t.Fatalf("Expected %d packages, got %d", len(samplePackages), len(resolver.Packages))
	}

	for i, pkg := range resolver.Packages {
		if pkg.Package != samplePackages[i].Package {
			t.Errorf("Expected package %s, got %s", samplePackages[i].Package, pkg.Package)
		}
		if pkg.Name != samplePackages[i].Name {
			t.Errorf("Expected name %s, got %s", samplePackages[i].Name, pkg.Name)
		}
		if pkg.Sdesc != samplePackages[i].Sdesc {
			t.Errorf("Expected short description %s, got %s", samplePackages[i].Sdesc, pkg.Sdesc)
		}
		if pkg.Ldesc != samplePackages[i].Ldesc {
			t.Errorf("Expected long description %s, got %s", samplePackages[i].Ldesc, pkg.Ldesc)
		}
		if pkg.Category != samplePackages[i].Category {
			t.Errorf("Expected category %s, got %s", samplePackages[i].Category, pkg.Category)
		}
		if len(pkg.Requires) != len(samplePackages[i].Requires) {
			t.Errorf("Expected requirements %v, got %v", samplePackages[i].Requires, pkg.Requires)
		}
	}
}

func TestSavePackageEntries(t *testing.T) {
	resolver := NewDependencyResolver()
	resolver.Packages = samplePackages

	resolver.SavePackageEntries(testFilePath)
	defer cleanupTestFile()

	content, err := ioutil.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	var data struct {
		Packages []PackageEntry `yaml:"packages"`
	}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		t.Fatalf("Error unmarshalling YAML: %v", err)
	}

	if len(data.Packages) != len(samplePackages) {
		t.Fatalf("Expected %d packages, got %d", len(samplePackages), len(data.Packages))
	}

	for i, pkg := range data.Packages {
		if pkg.Package != samplePackages[i].Package {
			t.Errorf("Expected package %s, got %s", samplePackages[i].Package, pkg.Package)
		}
		if pkg.Name != samplePackages[i].Name {
			t.Errorf("Expected name %s, got %s", samplePackages[i].Name, pkg.Name)
		}
		if pkg.Sdesc != samplePackages[i].Sdesc {
			t.Errorf("Expected short description %s, got %s", samplePackages[i].Sdesc, pkg.Sdesc)
		}
		if pkg.Ldesc != samplePackages[i].Ldesc {
			t.Errorf("Expected long description %s, got %s", samplePackages[i].Ldesc, pkg.Ldesc)
		}
		if pkg.Category != samplePackages[i].Category {
			t.Errorf("Expected category %s, got %s", samplePackages[i].Category, pkg.Category)
		}
		if len(pkg.Requires) != len(samplePackages[i].Requires) {
			t.Errorf("Expected requirements %v, got %v", samplePackages[i].Requires, pkg.Requires)
		}
	}
}
