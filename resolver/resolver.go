package resolver

import (
	"fmt"

	"github.com/spf13/afero"
)

type DependencyResolver struct {
	Fs                  afero.Fs
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

func NewDependencyResolver(fs afero.Fs) *DependencyResolver {
	return &DependencyResolver{
		Fs:                  fs,
		packageDependencies: make(map[string][]string),
		visitedPaths:        make(map[string]bool),
	}
}

func (dr *DependencyResolver) ShowPackageEntry(pkg string) {
	for _, entry := range dr.Packages {
		if entry.Package == pkg {
			fmt.Printf("Package: %s\nName: %s\nShort Description: %s\nLong Description: %s\nCategory: %s\nRequirements: %v\n",
				entry.Package, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
			return
		}
	}
	fmt.Printf("Package %s not found\n", pkg)
}
