package resolver

import (
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

func (dr *DependencyResolver) HandleRunCommand(resources []string) {
	logs := new(logs)

	for _, resName := range resources {
		for _, res := range dr.Resources {
			LogInfo("🔍 Resolving dependency " + res.Resource)

			if res.Resource == resName && res.Run != nil {
				for _, step := range res.Run {
					if step.Exec != "" {
						output, exitCode, err := exec.ExecuteCommand(step.Exec)

						if step.Name != "" {
							logs.addLogs(stepLog{targetRes: resName, command: step.Exec, res: res.Name, name: step.Name, message: output})
						}

						if err != nil {
							LogError("❌ Error executing command '"+step.Exec+"'", err)
						}

						if step.Expect != nil {
							expectations := expect.ProcessExpectations(step.Expect)
							if err := expect.CheckExpectations(output, exitCode, expectations); err != nil {
								LogError("❌ "+err.Error()+" for command '"+step.Exec+"'", err)
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

func (dr *DependencyResolver) HandleShowCommand(resources []string) {
	for _, res := range resources {
		dr.ShowResourceEntry(res)
	}
}

func (dr *DependencyResolver) HandleDependsCommand(resources []string) {
	for _, res := range resources {
		dr.Graph.ListDirectDependencies(res)
	}
}

func (dr *DependencyResolver) HandleRDependsCommand(resources []string) {
	for _, res := range resources {
		dr.Graph.ListReverseDependencies(res)
	}
}

func (dr *DependencyResolver) HandleSearchCommand(resources []string) {
	query := resources[0]
	keys := resources[1:]
	dr.FuzzySearch(query, keys)
}

func (dr *DependencyResolver) HandleCategoryCommand(resources []string) {
	if len(resources) == 0 {
		Println("Usage: kdeps category [categories...]")
		return
	}
	for _, entry := range dr.Resources {
		for _, category := range resources {
			if entry.Category == category {
				Println("📂 " + entry.Resource)
			}
		}
	}
}

func (dr *DependencyResolver) HandleTreeCommand(resources []string) {
	for _, res := range resources {
		dr.Graph.ListDependencyTree(res)
	}
}

func (dr *DependencyResolver) HandleTreeListCommand(resources []string) {
	for _, res := range resources {
		dr.Graph.ListDependencyTreeTopDown(res)
	}
}

func (dr *DependencyResolver) HandleIndexCommand() {
	for _, entry := range dr.Resources {
		PrintMessage("📦 Resource: %s\n📛 Name: %s\n📝 Short Description: %s\n📖 Long Description: %s\n🏷️ Category: %s\n🔗 Requirements: %v\n",
			entry.Resource, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
		Println("---")
	}
}
