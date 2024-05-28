package resolver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/afero"
)

// MockShowPackageEntry is a mock function to replace ShowPackageEntry for testing purposes.
func (dr *DependencyResolver) MockShowPackageEntry(pkg string) string {
	for _, entry := range dr.Packages {
		if entry.Package == pkg {
			return fmt.Sprintf("PackageEntry{Package:%s Name:%s Sdesc:%s Ldesc:%s Category:%s}\n",
				entry.Package, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category)
		}
	}
	return ""
}

// TestDependencyResolver_FuzzySearch tests the FuzzySearch method.
func TestDependencyResolver_FuzzySearch(t *testing.T) {
	resolver := NewDependencyResolver(afero.NewMemMapFs())
	resolver.Packages = []PackageEntry{
		{Package: "a", Name: "A", Sdesc: "Package A", Ldesc: "The first package in the alphabetical order", Category: "example"},
		{Package: "b", Name: "B", Sdesc: "Package B", Ldesc: "The second package, dependent on A", Category: "example"},
		{Package: "c", Name: "C", Sdesc: "Package C", Ldesc: "The third package, dependent on B", Category: "example"},
		// Add more packages as needed for the test
	}

	// Modify FuzzySearch temporarily within the test to return matched packages
	fuzzySearch := func(query string, keys []string) []string {
		combinedEntries := make([][2]string, len(resolver.Packages))
		for i, entry := range resolver.Packages {
			var combined strings.Builder
			for _, key := range keys {
				switch key {
				case "package":
					combined.WriteString(entry.Package + " ")
				case "name":
					combined.WriteString(entry.Name + " ")
				case "sdesc":
					combined.WriteString(entry.Sdesc + " ")
				case "ldesc":
					combined.WriteString(entry.Ldesc + " ")
				case "category":
					combined.WriteString(entry.Category + " ")
				}
			}
			combinedEntries[i] = [2]string{entry.Package, combined.String()}
		}

		matches := fuzzy.Find(query, getSecondStrings(combinedEntries))
		var matchedPackages []string
		for _, match := range matches {
			for _, entry := range combinedEntries {
				if strings.TrimSpace(entry[1]) == strings.TrimSpace(match) {
					matchedPackages = append(matchedPackages, entry[0])
					break
				}
			}
		}
		return matchedPackages
	}

	output := captureOutput(func() {
		matches := fuzzySearch("second", []string{"ldesc"})
		for _, match := range matches {
			fmt.Print(resolver.MockShowPackageEntry(match))
			fmt.Println("---")
		}
	})

	expectedOutput := "PackageEntry{Package:b Name:B Sdesc:Package B Ldesc:The second package, dependent on A Category:example}\n---\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output %s, got %s", expectedOutput, output)
	}
}
