package resolver

import (
	"fmt"
	"strings"
)

func (dr *DependencyResolver) ListDependencyTree(pkg string) {
	visited := make(map[string]bool)
	dr.listDependenciesRecursive(pkg, []string{}, visited)
}

func (dr *DependencyResolver) ListDependencyTreeTopDown(pkg string) {
	visited := make(map[string]bool)
	stack := dr.buildDependencyStack(pkg, visited)
	for _, node := range stack {
		fmt.Println(node)
	}
}

func (dr *DependencyResolver) listDependenciesRecursive(pkg string, path []string, visited map[string]bool) {
	if visited[pkg] {
		return
	}
	visited[pkg] = true
	path = append(path, pkg)
	deps, exists := dr.packageDependencies[pkg]
	if !exists || len(deps) == 0 {
		fmt.Println(strings.Join(path, " > "))
	} else {
		for _, dep := range deps {
			dr.listDependenciesRecursive(dep, path, visited)
		}
	}
	visited[pkg] = false
}

func (dr *DependencyResolver) buildDependencyStack(pkg string, visited map[string]bool) []string {
	if visited[pkg] {
		return nil
	}
	visited[pkg] = true

	var stack []string
	deps, exists := dr.packageDependencies[pkg]
	if exists {
		for _, dep := range deps {
			substack := dr.buildDependencyStack(dep, visited)
			stack = append(stack, substack...)
		}
	}
	stack = append(stack, pkg)
	return stack
}
