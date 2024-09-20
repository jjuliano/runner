package resolver

import (
	"fmt"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func (dr *DependencyResolver) FuzzySearch(query string, keys []string) error {
	if len(keys) == 0 {
		// If no keys are provided, search in all fields
		keys = []string{"id", "name", "desc", "category"}
	}

	combinedEntries := make([][2]string, len(dr.Resources))
	for i, entry := range dr.Resources {
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
	if len(matches) == 0 {
		LogErrorExit("No matches found for query: "+query, nil)
	}

	for _, match := range matches {
		for _, entry := range combinedEntries {
			if entry[1] == match {
				err := dr.ShowResourceEntry(entry[0])
				if err != nil {
					LogErrorExit("Failed to show resource entry: "+entry[0], err)
				}
				fmt.Println()
				break
			}
		}
	}
	return nil
}

func getSecondStrings(entries [][2]string) []string {
	secondStrings := make([]string, len(entries))
	for i, entry := range entries {
		secondStrings[i] = entry[1]
	}
	return secondStrings
}
