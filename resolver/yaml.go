package resolver

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func (dr *DependencyResolver) LoadResourceEntries(filePath string) error {
	data, err := afero.ReadFile(dr.Fs, filePath)
	if err != nil {
		return LogError("Error reading file "+filePath, err)
	}

	var fileResources struct {
		Resources []ResourceEntry `yaml:"resources"`
	}

	if err := yaml.Unmarshal(data, &fileResources); err != nil {
		return LogError("Error unmarshalling YAML data from file "+filePath, err)
	}

	dr.Resources = append(dr.Resources, fileResources.Resources...)
	for _, entry := range fileResources.Resources {
		dr.resourceDependencies[entry.Resource] = entry.Requires
	}
	return nil
}

func (dr *DependencyResolver) ShowResourceEntry(res string) error {
	for _, entry := range dr.Resources {
		if entry.Resource == res {
			PrintMessage("📦 Resource: %s\n📛 Name: %s\n📝 Short Description: %s\n📖 Long Description: %s\n🏷️  Category: %s\n🔗 Requirements: %v\n",
				entry.Resource, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
			return nil
		}
	}
	return LogError("Resource "+res+" not found", nil)
}

func (dr *DependencyResolver) SaveResourceEntries(filePath string) error {
	data := struct {
		Resources []ResourceEntry `yaml:"resources"`
	}{
		Resources: dr.Resources,
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		return LogError("Error marshalling YAML", err)
	}

	err = afero.WriteFile(dr.Fs, filePath, content, 0644)
	if err != nil {
		return LogError("Error writing file "+filePath, err)
	}
	return nil
}
