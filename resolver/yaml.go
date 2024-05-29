package resolver

import (
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func (dr *DependencyResolver) LoadPackageEntries(filePath string) {
	data, err := afero.ReadFile(dr.Fs, filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return
	}

	var filePackages struct {
		Packages []PackageEntry `yaml:"packages"`
	}

	if err := yaml.Unmarshal(data, &filePackages); err != nil {
		fmt.Printf("Error unmarshalling YAML data from file %s: %v\n", filePath, err)
		return
	}

	dr.Packages = append(dr.Packages, filePackages.Packages...)
	for _, entry := range filePackages.Packages {
		dr.packageDependencies[entry.Package] = entry.Requires
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

func (dr *DependencyResolver) SavePackageEntries(filePath string) {
	data := struct {
		Packages []PackageEntry `yaml:"packages"`
	}{
		Packages: dr.Packages,
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling YAML:", err)
		return
	}

	err = afero.WriteFile(dr.Fs, filePath, content, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
	}
}
