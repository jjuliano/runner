package resolver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/jjuliano/runner/pkg/runnerexec"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/afero"
)

// MockShowResourceEntry is a mock function to replace ShowResourceEntry for testing purposes.
func (dr *DependencyResolver) MockShowResourceEntry(res string) string {
	for _, entry := range dr.Resources {
		if entry.Id == res {
			return fmt.Sprintf("ResourceNodeEntry{Id:%s Name:%s Desc:%s Category:%s}\n",
				entry.Id, entry.Name, entry.Desc, entry.Category)
		}
	}
	return ""
}

// TestDependencyResolver_FuzzySearch tests the FuzzySearch method.
func TestDependencyResolver_FuzzySearch(t *testing.T) {
	logger := log.New(nil)
	session, err := runnerexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()
	resolver, err := NewGraphResolver(afero.NewMemMapFs(), logger, "", session)
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	resolver.Resources = []ResourceNodeEntry{
		{Id: "a", Name: "A", Desc: "The first resource in the alphabetical order", Category: "example"},
		{Id: "b", Name: "B", Desc: "The second resource, dependent on A", Category: "example"},
		{Id: "c", Name: "C", Desc: "The third resource, dependent on B", Category: "example"},
	}

	// Modify FuzzySearch temporarily within the test to return matched resources
	fuzzySearch := func(query string, keys []string) []string {
		combinedEntries := make([][2]string, len(resolver.Resources))
		for i, entry := range resolver.Resources {
			var combined strings.Builder
			for _, key := range keys {
				switch key {
				case "id":
					combined.WriteString(entry.Id + " ")
				case "name":
					combined.WriteString(entry.Name + " ")
				case "desc":
					combined.WriteString(entry.Desc + " ")
				case "category":
					combined.WriteString(entry.Category + " ")
				}
			}
			combinedEntries[i] = [2]string{entry.Id, combined.String()}
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
		matches := fuzzySearch("second", []string{"desc"})
		for _, match := range matches {
			fmt.Print(resolver.MockShowResourceEntry(match))
			fmt.Println()
		}
	})

	expectedOutput := "ResourceNodeEntry{Id:b Name:B Desc:The second resource, dependent on A Category:example}\n\n"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output %s, got %s", expectedOutput, output)
	}
}
