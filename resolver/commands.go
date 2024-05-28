package resolver

import "fmt"

func (dr *DependencyResolver) HandleAddCommand(packages []string) {
	if len(packages) < 6 {
		fmt.Println("Usage: script add package_name name short_desc long_desc category [requirements...]")
		return
	}
	packageName, name, shortDesc, longDesc, category := packages[0], packages[1], packages[2], packages[3], packages[4]
	requirements := packages[5:]
	dr.AddPackageEntry("setup.yml", packageName, name, shortDesc, longDesc, category, requirements)
}

func (dr *DependencyResolver) HandleUpdateCommand(packages []string) {
	if len(packages) < 6 {
		fmt.Println("Usage: script update package_name name short_desc long_desc category [requirements...]")
		return
	}
	packageName, name, shortDesc, longDesc, category := packages[0], packages[1], packages[2], packages[3], packages[4]
	requirements := packages[5:]
	dr.UpdatePackageEntry("setup.yml", packageName, name, shortDesc, longDesc, category, requirements)
}

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
	if len(packages) < 2 {
		fmt.Println("Usage: script search query [keys...]")
		return
	}
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
