package resolver

import (
	"fmt"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func (dr *DependencyResolver) FuzzySearch(query string, keys []string) {
	if len(keys) == 0 {
		// If no keys are provided, search in all fields
		keys = []string{"resource", "name", "sdesc", "ldesc", "category"}
	}

	combinedEntries := make([][2]string, len(dr.Resources))
	for i, entry := range dr.Resources {
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
	for _, match := range matches {
		for _, entry := range combinedEntries {
			if entry[1] == match {
				dr.ShowResourceEntry(entry[0])
				fmt.Println("---")
				break
			}
		}
	}
}
