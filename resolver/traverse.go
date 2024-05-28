package resolver

import (
	"fmt"
	"strings"
)

func (dr *DependencyResolver) ListDirectDependencies(pkg string) {
	dr.dependencyGraph = []string{}
	dr.visitedPaths = make(map[string]bool)
	visited := make(map[string]bool)
	dr.traverseDependencyGraph(pkg, dr.packageDependencies, visited)
}

func (dr *DependencyResolver) ListReverseDependencies(pkg string) {
	inverted := dr.invertDependencies()
	dr.dependencyGraph = []string{}
	dr.visitedPaths = make(map[string]bool)
	visited := make(map[string]bool)
	dr.traverseDependencyGraph(pkg, inverted, visited)
}

func (dr *DependencyResolver) traverseDependencyGraph(pkg string, dependencies map[string][]string, visited map[string]bool) {
	if visited[pkg] {
		return
	}
	visited[pkg] = true
	dr.dependencyGraph = append(dr.dependencyGraph, pkg)

	currentPath := strings.Join(dr.dependencyGraph, " > ")
	if dr.visitedPaths[currentPath] {
		dr.dependencyGraph = dr.dependencyGraph[:len(dr.dependencyGraph)-1]
		return
	}
	dr.visitedPaths[currentPath] = true
	fmt.Println(currentPath)

	if deps, exists := dependencies[pkg]; exists {
		for _, dep := range deps {
			dr.traverseDependencyGraph(dep, dependencies, visited)
		}
	}
	dr.dependencyGraph = dr.dependencyGraph[:len(dr.dependencyGraph)-1]
}

func (dr *DependencyResolver) invertDependencies() map[string][]string {
	inverted := make(map[string][]string)
	for pkg, deps := range dr.packageDependencies {
		for _, dep := range deps {
			inverted[dep] = append(inverted[dep], pkg)
		}
	}
	return inverted
}
