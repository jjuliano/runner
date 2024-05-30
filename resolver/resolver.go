package resolver

import (
	"github.com/charmbracelet/log"
	graph "github.com/kdeps/kartographer/graph"
	"github.com/spf13/afero"
)

type DependencyResolver struct {
	Fs                  afero.Fs
	Packages            []PackageEntry
	packageDependencies map[string][]string
	dependencyGraph     []string
	visitedPaths        map[string]bool
	logger              *log.Logger
	Graph               *graph.DependencyGraph
}

type RunStep struct {
	Name   string `yaml:"name"`
	Exec   string `yaml:"exec"`
	Expect string `yaml:"expect"`
}

type PackageEntry struct {
	Package  string    `yaml:"package"`
	Name     string    `yaml:"name"`
	Sdesc    string    `yaml:"sdesc"`
	Ldesc    string    `yaml:"ldesc"`
	Category string    `yaml:"category"`
	Requires []string  `yaml:"requires"`
	Run      []RunStep `yaml:"run"`
}

func NewDependencyResolver(fs afero.Fs, logger *log.Logger) *DependencyResolver {
	dr := &DependencyResolver{
		Fs:                  fs,
		packageDependencies: make(map[string][]string),
		visitedPaths:        make(map[string]bool),
		logger:              logger,
	}
	dr.Graph = graph.NewDependencyGraph(fs, logger, dr.packageDependencies)
	return dr
}
