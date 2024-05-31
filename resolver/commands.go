package resolver

import (
	"net/http"
	"strings"

	"github.com/kdeps/plugins/exec"
	"github.com/kdeps/plugins/expect"
)

type logs []stepLog

type stepLog struct {
	name      string
	message   string
	res       string
	command   string
	targetRes string
}

func (m *logs) addLogs(entry stepLog) {
	*m = append(*m, entry)
}

func (m logs) getLogs() []stepLog {
	var logEntries []stepLog
	for _, val := range m {
		logEntries = append(logEntries, stepLog{targetRes: val.targetRes, command: val.command, res: val.res, name: val.name, message: val.message})
	}
	return logEntries
}

func formatLogEntry(entry stepLog) string {
	return strings.Join([]string{
		"\n----------------------------------------------------------\n",
		"📦 Resource: " + entry.res,
		"🔄 Step: " + entry.name,
		"💻 Command: " + entry.command,
		"\n" + entry.message,
		"----------------------------------------------------------",
	}, "\n")
}

func (dr *DependencyResolver) HandleRunCommand(resources []string) error {
	logs := new(logs)
	visited := make(map[string]bool)
	client := &http.Client{}

	for _, resName := range resources {
		stack := dr.Graph.BuildDependencyStack(resName, visited)

		for _, resNode := range stack {
			for _, res := range dr.Resources {
				if res.Resource == resNode {
					LogInfo("🔍 Resolving dependency " + resNode)
					if res.Run != nil {

						for _, step := range res.Run {
							if step.Exec != "" {
								output, exitCode, err := exec.ExecuteCommand(step.Exec)

								if step.Name != "" {
									logs.addLogs(stepLog{targetRes: resNode, command: step.Exec, res: res.Name, name: step.Name, message: output})
								}

								if err != nil {
									return LogError("Error executing command '"+step.Exec+"'", err)
								}

								if step.Expect != nil {
									expectations := expect.ProcessExpectations(step.Expect)
									if err := expect.CheckExpectations(output, exitCode, expectations, client); err != nil {
										return LogError("Expectation check failed for command '"+step.Exec+"'", err)
									}
								}
							}
						}
					}
				}

				for _, val := range logs.getLogs() {
					if val.name != "" {
						formattedLog := formatLogEntry(val)
						LogInfo("🏃 Running " + val.name + "... " + formattedLog)
					}
				}
			}
		}
	}
	return nil
}

func (dr *DependencyResolver) HandleShowCommand(resources []string) error {
	for _, res := range resources {
		if err := dr.ShowResourceEntry(res); err != nil {
			return LogError("Error showing resource entry "+res, err)
		}
	}
	return nil
}

func (dr *DependencyResolver) HandleDependsCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListDirectDependencies(res) // TODO: Return error on kartographer plugin
	}
	return nil
}

func (dr *DependencyResolver) HandleRDependsCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListReverseDependencies(res) // TODO: Return error on kartographer plugin
	}
	return nil
}

func (dr *DependencyResolver) HandleSearchCommand(resources []string) error {
	query := resources[0]
	keys := resources[1:]
	return dr.FuzzySearch(query, keys)
}

func (dr *DependencyResolver) HandleCategoryCommand(resources []string) error {
	if len(resources) == 0 {
		Println("Usage: kdeps category [categories...]")
		return nil
	}
	for _, entry := range dr.Resources {
		for _, category := range resources {
			if entry.Category == category {
				Println("📂 " + entry.Resource)
			}
		}
	}
	return nil
}

func (dr *DependencyResolver) HandleTreeCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListDependencyTree(res) // TODO: Return error on kartographer plugin
	}
	return nil
}

func (dr *DependencyResolver) HandleTreeListCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListDependencyTreeTopDown(res) // TODO: Return error on kartographer plugin
	}
	return nil
}

func (dr *DependencyResolver) HandleIndexCommand() error {
	for _, entry := range dr.Resources {
		PrintMessage("📦 Resource: %s\n📛 Name: %s\n📝 Short Description: %s\n📖 Long Description: %s\n🏷️  Category: %s\n🔗 Requirements: %v\n",
			entry.Resource, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
		Println("---")
	}
	return nil
}
