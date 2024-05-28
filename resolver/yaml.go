package resolver

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func (dr *DependencyResolver) LoadPackageEntries(file string) {
	content, err := ioutil.ReadFile(file)
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
	for _, entry := range dr.Packages {
		dr.packageDependencies[entry.Package] = entry.Requires
	}
}

func (dr *DependencyResolver) SavePackageEntries(file string) {
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

	err = ioutil.WriteFile(file, content, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
	}
}
