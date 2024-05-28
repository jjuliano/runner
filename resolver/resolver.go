package resolver

import (
	"fmt"
)

type DependencyResolver struct {
	Packages            []PackageEntry
	packageDependencies map[string][]string
	dependencyGraph     []string
	visitedPaths        map[string]bool
}

type PackageEntry struct {
	Package  string   `yaml:"package"`
	Name     string   `yaml:"name"`
	Sdesc    string   `yaml:"sdesc"`
	Ldesc    string   `yaml:"ldesc"`
	Category string   `yaml:"category"`
	Requires []string `yaml:"requires"`
}

func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{
		Packages:            []PackageEntry{},
		packageDependencies: make(map[string][]string),
		dependencyGraph:     []string{},
		visitedPaths:        make(map[string]bool),
	}
}

func (dr *DependencyResolver) ShowPackageEntry(packageName string) {
	for _, entry := range dr.Packages {
		if entry.Package == packageName {
			fmt.Printf("Package: %s\nName: %s\nShort Description: %s\nLong Description: %s\nCategory: %s\nRequirements: %s\n",
				entry.Package, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
			return
		}
	}
	fmt.Printf("Package %s not found.\n", packageName)
}
