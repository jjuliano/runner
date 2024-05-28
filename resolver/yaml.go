package resolver

import (
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func (dr *DependencyResolver) LoadPackageEntries(filePath string) {
	content, err := afero.ReadFile(dr.Fs, filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var data struct {
		Packages []PackageEntry `yaml:"packages"`
	}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		fmt.Println("Error unmarshalling YAML:", err)
		return
	}

	dr.Packages = data.Packages
	for _, pkg := range dr.Packages {
		dr.packageDependencies[pkg.Package] = pkg.Requires
	}
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
