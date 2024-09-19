package resolver

import (
    "fmt"

    "github.com/charmbracelet/log"
    "github.com/kdeps/kartographer/graph"
    "github.com/jjuliano/runner/pkg/kdepexec"
    "github.com/spf13/afero"
)

type DependencyResolver struct {
    Fs                   afero.Fs
    Resources            []ResourceNodeEntry
    ResourceDependencies map[string][]string
    DependencyGraph      []string
    VisitedPaths         map[string]bool
    Logger               *log.Logger
    Graph                *graph.DependencyGraph
    WorkDir              string
    ShellSession         *kdepexec.ShellSession
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
    Input string `yaml:"input,omitempty"`
    File  string `yaml:"file,omitempty"`
}

type StepKey struct {
    name string
    node string
}

type ResourceNodeEntry struct {
    Id       string    `yaml:"id"`
    Name     string    `yaml:"name"`
    Desc     string    `yaml:"desc"`
    Category string    `yaml:"category"`
    Requires []string  `yaml:"requires"`
    Run      []RunStep `yaml:"run"`
}

func NewGraphResolver(fs afero.Fs, logger *log.Logger, workDir string, shellSession *kdepexec.ShellSession) (*DependencyResolver, error) {
    dependencyResolver := &DependencyResolver{
        Fs:                   fs,
        ResourceDependencies: make(map[string][]string),
        VisitedPaths:         make(map[string]bool),
        Logger:               logger,
        WorkDir:              workDir,
        ShellSession:         shellSession,
    }

    dependencyResolver.Graph = graph.NewDependencyGraph(fs, logger, dependencyResolver.ResourceDependencies)
    if dependencyResolver.Graph == nil {
        return nil, fmt.Errorf("failed to initialize dependency graph")
    }
    return dependencyResolver, nil
}
