package resolver

import "fmt"

func (dr *DependencyResolver) HandleShowCommand(packages []string) {
	for _, pkg := range packages {
		dr.ShowPackageEntry(pkg)
	}
}

func (dr *DependencyResolver) HandleDependsCommand(packages []string) {
	for _, pkg := range packages {
		dr.ListDirectDependencies(pkg)
	}
}

func (dr *DependencyResolver) HandleRDependsCommand(packages []string) {
	for _, pkg := range packages {
		dr.ListReverseDependencies(pkg)
	}
}

func (dr *DependencyResolver) HandleSearchCommand(packages []string) {
	query := packages[0]
	keys := packages[1:]
	dr.FuzzySearch(query, keys)
}

func (dr *DependencyResolver) HandleCategoryCommand(packages []string) {
	if len(packages) == 0 {
		fmt.Println("Usage: script category [categories...]")
		return
	}
	for _, entry := range dr.Packages {
		for _, category := range packages {
			if entry.Category == category {
				fmt.Println(entry.Package)
			}
		}
	}
}

func (dr *DependencyResolver) HandleTreeCommand(packages []string) {
	for _, pkg := range packages {
		dr.ListDependencyTree(pkg)
	}
}

func (dr *DependencyResolver) HandleTreeListCommand(packages []string) {
	for _, pkg := range packages {
		dr.ListDependencyTreeTopDown(pkg)
	}
}
