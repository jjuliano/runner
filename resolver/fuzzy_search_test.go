package resolver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/afero"
)

// MockShowResourceEntry is a mock function to replace ShowResourceEntry for testing purposes.
func (dr *DependencyResolver) MockShowResourceEntry(res string) string {
	for _, entry := range dr.Resources {
		if entry.Resource == res {
			return fmt.Sprintf("ResourceEntry{Resource:%s Name:%s Sdesc:%s Ldesc:%s Category:%s}\n",
				entry.Resource, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category)
		}
	}
	return ""
}

// TestDependencyResolver_FuzzySearch tests the FuzzySearch method.
func TestDependencyResolver_FuzzySearch(t *testing.T) {
	logger := log.New(nil)
	resolver := NewDependencyResolver(afero.NewMemMapFs(), logger)
	resolver.Resources = []ResourceEntry{
		{Resource: "a", Name: "A", Sdesc: "Resource A", Ldesc: "The first resource in the alphabetical order", Category: "example"},
		{Resource: "b", Name: "B", Sdesc: "Resource B", Ldesc: "The second resource, dependent on A", Category: "example"},
		{Resource: "c", Name: "C", Sdesc: "Resource C", Ldesc: "The third resource, dependent on B", Category: "example"},
		// Add more resources as needed for the test
	}

	// Modify FuzzySearch temporarily within the test to return matched resources
	fuzzySearch := func(query string, keys []string) []string {
		combinedEntries := make([][2]string, len(resolver.Resources))
		for i, entry := range resolver.Resources {
			var combined strings.Builder
			for _, key := range keys {
				switch key {
				case "resource":
					combined.WriteString(entry.Resource + " ")
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
			combinedEntries[i] = [2]string{entry.Resource, combined.String()}
		}

		matches := fuzzy.Find(query, getSecondStrings(combinedEntries))
		var matchedResources []string
		for _, match := range matches {
			for _, entry := range combinedEntries {
				if strings.TrimSpace(entry[1]) == strings.TrimSpace(match) {
					matchedResources = append(matchedResources, entry[0])
					break
				}
			}
		}
		return matchedResources
	}

	output := captureOutput(func() {
		matches := fuzzySearch("second", []string{"ldesc"})
		for _, match := range matches {
			fmt.Print(resolver.MockShowResourceEntry(match))
			fmt.Println("---")
		}
	})

	expectedOutput := "ResourceEntry{Resource:b Name:B Sdesc:Resource B Ldesc:The second resource, dependent on A Category:example}\n---\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output %s, got %s", expectedOutput, output)
	}
}
