package resolver

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

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

func (m *logs) addLogs(entry stepLog, logChan chan<- stepLog) {
	*m = append(*m, entry)
	logChan <- entry // Send log entry to the channel
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

func (dr *DependencyResolver) processSteps(initSteps []interface{}, client *http.Client) error {
	for _, init := range initSteps {
		if err := processElement(init, client); err != nil {
			return err
		}
	}
	return nil
}

func processElement(init interface{}, client *http.Client) error {
	switch val := init.(type) {
	case string:
		if isValidCheckPrefix(val) {
			return expect.CheckExpectations("", 0, []string{val}, client)
		}
	case map[interface{}]interface{}:
		if expectVal, exists := val["expect"]; exists {
			switch expectType := expectVal.(type) {
			case []interface{}:
				// Convert expectType to []string
				strs := make([]string, len(expectType))
				for i, v := range expectType {
					if s, ok := v.(string); ok {
						strs[i] = s
					} else {
						return fmt.Errorf("unsupported expect value: %v", v)
					}
				}
				return expect.CheckExpectations("", 0, strs, client)
			case string:
				if isValidCheckPrefix(expectType) {
					return expect.CheckExpectations("", 0, []string{expectType}, client)
				}
				return fmt.Errorf("unsupported expect value: %v", expectVal)
			default:
				return fmt.Errorf("unsupported expect value type: %T", expectVal)
			}
		}
	}
	return nil
}

func isValidCheckPrefix(s string) bool {
	return strings.HasPrefix(s, "ENV:") ||
		strings.HasPrefix(s, "FILE:") ||
		strings.HasPrefix(s, "DIR:") ||
		strings.HasPrefix(s, "URL:") ||
		strings.HasPrefix(s, "!") ||
		(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\""))
}

// Execute the command and handle the result
func executeAndLogCommand(step RunStep, resName, resNode string, logs *logs, logChan chan<- stepLog, client *http.Client) error {
	result := <-exec.ExecuteCommand(step.Exec)

	if step.Name != "" {
		logs.addLogs(stepLog{
			targetRes: resNode,
			command:   step.Exec,
			res:       resName,
			name:      step.Name,
			message:   result.Output,
		}, logChan)
	}

	if result.Err != nil {
		return result.Err
	}

	if step.Expect != nil {
		expectations := expect.ProcessExpectations(step.Expect)
		if err := expect.CheckExpectations(result.Output, result.ExitCode, expectations, client); err != nil {
			return err
		}
	}
	return nil
}

// Main function to handle run commands
func (dr *DependencyResolver) HandleRunCommand(resources []string) error {
	logs := new(logs)
	visited := make(map[string]bool)
	client := &http.Client{}
	logChan := make(chan stepLog)

	go func() {
		for logEntry := range logChan {
			if logEntry.name != "" {
				formattedLog := formatLogEntry(logEntry)
				LogInfo("🏃 Running " + logEntry.name + "... " + formattedLog)
			}
		}
	}()

	defer close(logChan) // Ensure the channel is closed when done

	for _, resName := range resources {
		stack := dr.Graph.BuildDependencyStack(resName, visited)
		for _, resNode := range stack {
			for _, res := range dr.Resources {
				if res.Resource == resNode {
					LogInfo("🔍 Resolving dependency " + resNode)
					if res.Run != nil {
						var wg sync.WaitGroup
						skipResults := make(chan bool, len(res.Run))

						// Handle skip steps with expect
						for _, step := range res.Run {
							wg.Add(1)
							go func(step RunStep) {
								defer wg.Done()
								if skipSteps, ok := step.Skip.([]interface{}); ok {
									if err := dr.processSteps(skipSteps, client); err != nil {
										LogError("Skip expectation failed for resource '"+resNode+"' step '"+step.Name+"'", err)
										skipResults <- false
									} else {
										LogInfo("Skip step succeeded for resource '" + resNode + "' step '" + step.Name + "'")
										skipResults <- true
									}
								} else {
									skipResults <- false
								}
							}(step)
						}

						// Wait for all skip checks to complete
						wg.Wait()
						close(skipResults)

						// Determine if any step should be skipped
						skip := false
						for result := range skipResults {
							if result {
								skip = true
								break
							}
						}

						// Handle the rest of the steps only if not skipped
						if !skip {
							// Handle check steps with expect
							for _, step := range res.Run {
								if checkSteps, ok := step.Check.([]interface{}); ok {
									if err := dr.processSteps(checkSteps, client); err != nil {
										LogError("Check expectation failed for resource '"+resNode+"' step '"+step.Name+"'", err)
									}
								}
							}

							// Handle exec steps
							for _, step := range res.Run {
								if step.Exec != "" {
									if err := executeAndLogCommand(step, res.Name, resNode, logs, logChan, client); err != nil {
										LogError("Error executing command '"+resNode+"' step '"+step.Name+"'", err)
									}
								}
							}
						}
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
