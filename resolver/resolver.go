package resolver

import (
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
