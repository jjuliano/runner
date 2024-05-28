package resolver

import "fmt"

func (dr *DependencyResolver) AddPackageEntry(file, packageName, name, shortDesc, longDesc, category string, requirements []string) {
	newEntry := PackageEntry{
		Package:  packageName,
		Name:     name,
		Sdesc:    shortDesc,
		Ldesc:    longDesc,
		Category: category,
		Requires: requirements,
	}

	dr.Packages = append(dr.Packages, newEntry)
	dr.packageDependencies[packageName] = requirements
	dr.SavePackageEntries(file)
}

func (dr *DependencyResolver) UpdatePackageEntry(file, packageName, name, shortDesc, longDesc, category string, requirements []string) {
	for i, entry := range dr.Packages {
		if entry.Package == packageName {
			dr.Packages[i] = PackageEntry{
				Package:  packageName,
				Name:     name,
				Sdesc:    shortDesc,
				Ldesc:    longDesc,
				Category: category,
				Requires: requirements,
			}
			dr.packageDependencies[packageName] = requirements
			break
		}
	}
	dr.SavePackageEntries(file)
	fmt.Printf("Package %s updated successfully.\n", packageName)
}
