package resolver

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func (dr *DependencyResolver) LoadResourceEntries(filePath string) {
	data, err := afero.ReadFile(dr.Fs, filePath)
	if err != nil {
		PrintMessage("Error reading file %s: %v\n", filePath, err)
		return
	}

	var fileResources struct {
		Resources []ResourceEntry `yaml:"resources"`
	}

	if err := yaml.Unmarshal(data, &fileResources); err != nil {
		PrintMessage("Error unmarshalling YAML data from file %s: %v\n", filePath, err)
		return
	}

	dr.Resources = append(dr.Resources, fileResources.Resources...)
	for _, entry := range fileResources.Resources {
		dr.resourceDependencies[entry.Resource] = entry.Requires
	}
}

func (dr *DependencyResolver) ShowResourceEntry(res string) {
	for _, entry := range dr.Resources {
		if entry.Resource == res {
			PrintMessage("Resource: %s\nName: %s\nShort Description: %s\nLong Description: %s\nCategory: %s\nRequirements: %v\n",
				entry.Resource, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
			return
		}
	}
	PrintMessage("Resource %s not found\n", res)
}

func (dr *DependencyResolver) SaveResourceEntries(filePath string) {
	data := struct {
		Resources []ResourceEntry `yaml:"resources"`
	}{
		Resources: dr.Resources,
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		Println("Error marshalling YAML:", err)
		return
	}

	err = afero.WriteFile(dr.Fs, filePath, content, 0644)
	if err != nil {
		Println("Error writing file:", err)
	}
}
