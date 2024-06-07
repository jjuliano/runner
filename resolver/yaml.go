package resolver

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func (dr *DependencyResolver) LoadResourceEntries(filePath string) error {
	data, err := afero.ReadFile(dr.Fs, filePath)
	if err != nil {
		LogErrorExit("Error reading file "+filePath, err)
	}

	var fileResources struct {
		Resources []ResourceNodeEntry `yaml:"resources"`
	}

	if err := yaml.Unmarshal(data, &fileResources); err != nil {
		LogErrorExit("Error unmarshalling YAML data from file "+filePath, err)
	}

	dr.Resources = append(dr.Resources, fileResources.Resources...)
	for _, entry := range fileResources.Resources {
		dr.ResourceDependencies[entry.Id] = entry.Requires
	}
	return nil
}

func (dr *DependencyResolver) ShowResourceEntry(res string) error {
	for _, entry := range dr.Resources {
		if entry.Id == res {
			PrintMessage("📦 Id: %s\n📛 Name: %s\n📝 Description: %s\n🏷️  Category: %s\n🔗 Requirements: %v\n",
				entry.Id, entry.Name, entry.Desc, entry.Category, entry.Requires)
			return nil
		}
	}
	LogErrorExit("Id "+res+" not found", nil)
	return nil
}

func (dr *DependencyResolver) SaveResourceEntries(filePath string) error {
	data := struct {
		Resources []ResourceNodeEntry `yaml:"resources"`
	}{
		Resources: dr.Resources,
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		LogErrorExit("Error marshalling YAML", err)
	}

	err = afero.WriteFile(dr.Fs, filePath, content, 0644)
	if err != nil {
		LogErrorExit("Error writing file "+filePath, err)
	}
	return nil
}
