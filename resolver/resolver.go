package resolver

import (
	"fmt"

	"github.com/charmbracelet/log"
	graph "github.com/kdeps/kartographer/graph"
	"github.com/spf13/afero"
)

type DependencyResolver struct {
	Fs                   afero.Fs
	Resources            []ResourceEntry
	resourceDependencies map[string][]string
	dependencyGraph      []string
	visitedPaths         map[string]bool
	logger               *log.Logger
	Graph                *graph.DependencyGraph
	WorkDir              string
}

type RunStep struct {
	Name   string      `yaml:"name"`
	Exec   string      `yaml:"exec"`
	Skip   interface{} `yaml:"skip"`
	Check  interface{} `yaml:"check"`
	Expect interface{} `yaml:"expect"`
	Env    []EnvVar    `yaml:"env"`
}

type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value,omitempty"`
	Exec  string `yaml:"exec,omitempty"`
}

type StepKey struct {
	name string
	node string
}

type ResourceEntry struct {
	Resource string    `yaml:"resource"`
	Name     string    `yaml:"name"`
	Sdesc    string    `yaml:"sdesc"`
	Ldesc    string    `yaml:"ldesc"`
	Category string    `yaml:"category"`
	Requires []string  `yaml:"requires"`
	Run      []RunStep `yaml:"run"`
}

func NewDependencyResolver(fs afero.Fs, logger *log.Logger, workDir string) (*DependencyResolver, error) {
	dr := &DependencyResolver{
		Fs:                   fs,
		resourceDependencies: make(map[string][]string),
		visitedPaths:         make(map[string]bool),
		logger:               logger,
		WorkDir:              workDir,
	}

	dr.Graph = graph.NewDependencyGraph(fs, logger, dr.resourceDependencies)
	if dr.Graph == nil {
		return nil, fmt.Errorf("failed to initialize dependency graph")
	}
	return dr, nil
}
