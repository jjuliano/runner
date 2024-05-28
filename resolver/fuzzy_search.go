package resolver

import (
	"fmt"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func (dr *DependencyResolver) FuzzySearch(query string, keys []string) {
	combinedEntries := make([][2]string, len(dr.Packages))
	for i, entry := range dr.Packages {
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
	for _, match := range matches {
		for _, entry := range combinedEntries {
			if entry[1] == match {
				dr.ShowPackageEntry(entry[0])
				fmt.Println("---")
				break
			}
		}
	}
}
