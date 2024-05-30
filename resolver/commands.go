package resolver

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type logs []stepLog

type stepLog struct {
	name      string
	message   string
	pkg       string
	command   string
	targetPkg string
}

func (m *logs) addLogs(entry stepLog) {
	*m = append(*m, entry)
}

func (m logs) getLogs() []stepLog {
	var logEntries []stepLog
	for _, val := range m {
		logEntries = append(logEntries, stepLog{targetPkg: val.targetPkg, command: val.command, pkg: val.pkg, name: val.name, message: val.message})
	}
	return logEntries
}

func executeCommand(execCmd string) (string, error) {
	cmd := exec.Command("sh", "-c", execCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func formatLogEntry(entry stepLog) string {
	return strings.Join([]string{
		"\n----------------------------------------------------------\n",
		"📦 Package: " + entry.pkg,
		"🔄 Step: " + entry.name,
		"💻 Command: " + entry.command,
		"\n" + entry.message,
		"----------------------------------------------------------",
	}, "\n")
}

func (dr *DependencyResolver) HandleRunCommand(packages []string) {
	logs := new(logs)

	for _, pkgName := range packages {
		for _, pkg := range dr.Packages {
			dr.logger.Info(fmt.Sprintf("🔍 Resolving dependency %s", pkg.Package))

			if pkg.Package == pkgName && pkg.Run != nil {
				for _, step := range pkg.Run {
					if step.Exec != "" {
						output, err := executeCommand(step.Exec)

						if step.Name != "" {
							logs.addLogs(stepLog{targetPkg: pkgName, command: step.Exec, pkg: pkg.Name, name: step.Name, message: output})
						}

						if err != nil {
							dr.logger.Errorf("❌ Error executing command '%s': %s\n", step.Exec, err)
							os.Exit(1)
						}

						if step.Expect != "" {
							if !strings.Contains(strings.ToLower(output), strings.ToLower(step.Expect)) {
								dr.logger.Errorf("❌ Expected '%s' not found in output of '%s'\n", step.Expect, step.Exec)
								os.Exit(1)
							}
						}
					}
				}
			}

			for _, val := range logs.getLogs() {
				if val.name != "" {
					formattedLog := formatLogEntry(val)
					dr.logger.Printf("🏃 Running %s... %s", val.name, formattedLog)
				}
			}
		}
	}
}

func (dr *DependencyResolver) HandleShowCommand(packages []string) {
	for _, pkg := range packages {
		dr.ShowPackageEntry(pkg)
	}
}

func (dr *DependencyResolver) HandleDependsCommand(packages []string) {
	for _, pkg := range packages {
		dr.Graph.ListDirectDependencies(pkg)
	}
}

func (dr *DependencyResolver) HandleRDependsCommand(packages []string) {
	for _, pkg := range packages {
		dr.Graph.ListReverseDependencies(pkg)
	}
}

func (dr *DependencyResolver) HandleSearchCommand(packages []string) {
	query := packages[0]
	keys := packages[1:]
	dr.FuzzySearch(query, keys)
}

func (dr *DependencyResolver) HandleCategoryCommand(packages []string) {
	if len(packages) == 0 {
		fmt.Println("Usage: kdeps category [categories...]")
		return
	}
	for _, entry := range dr.Packages {
		for _, category := range packages {
			if entry.Category == category {
				fmt.Println("📂 " + entry.Package)
			}
		}
	}
}

func (dr *DependencyResolver) HandleTreeCommand(packages []string) {
	for _, pkg := range packages {
		dr.Graph.ListDependencyTree(pkg)
	}
}

func (dr *DependencyResolver) HandleTreeListCommand(packages []string) {
	for _, pkg := range packages {
		dr.Graph.ListDependencyTreeTopDown(pkg)
	}
}

func (dr *DependencyResolver) HandleIndexCommand() {
	for _, entry := range dr.Packages {
		fmt.Printf("📦 Package: %s\n📛 Name: %s\n📝 Short Description: %s\n📖 Long Description: %s\n🏷️ Category: %s\n🔗 Requirements: %v\n",
			entry.Package, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
		fmt.Println("---")
	}
}
