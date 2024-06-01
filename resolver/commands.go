package resolver

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kdeps/plugins/exec"
	"github.com/kdeps/plugins/expect"
)

// stepLog represents the structure of a log entry for a step.
type stepLog struct {
	name      string
	message   string
	res       string
	command   string
	targetRes string
}

// logs manages the logging mechanism with synchronization.
type logs struct {
	mu      sync.Mutex
	entries []stepLog
	closed  bool
}

// add adds a new log entry to the logs.
func (m *logs) add(entry stepLog) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	fmt.Println(formatLogEntry(entry)) // Use formatLogEntry for structured logging
	m.entries = append(m.entries, entry)
}

// close closes the log after all goroutines are done.
func (m *logs) close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
}

// getAll retrieves all log entries.
func (m *logs) getAll() []stepLog {
	m.mu.Lock()
	defer m.mu.Unlock()
	logEntries := make([]stepLog, len(m.entries))
	copy(logEntries, m.entries)
	return logEntries
}

// formatLogEntry formats a log entry into a string.
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

// processSteps processes each step by executing the relevant checks.
func (dr *DependencyResolver) processSteps(steps []interface{}, stepType, resNode string, client *http.Client) error {
	for _, step := range steps {
		LogInfo(fmt.Sprintf("Processing '%s' step: '%v' - '%s'", stepType, step, resNode))
		if err := processElement(step, client); err != nil {
			return err
		}
	}
	return nil
}

// processElement processes an individual step element based on its type.
func processElement(element interface{}, client *http.Client) error {
	switch val := element.(type) {
	case string:
		if isValidCheckPrefix(val) {
			return expect.CheckExpectations("", 0, []string{val}, client)
		}
	case map[interface{}]interface{}:
		if expectVal, exists := val["expect"]; exists {
			switch ev := expectVal.(type) {
			case []interface{}:
				return checkExpectations(ev, client)
			case string:
				if isValidCheckPrefix(ev) {
					return expect.CheckExpectations("", 0, []string{ev}, client)
				}
				return fmt.Errorf("unsupported expect value: %v", expectVal)
			default:
				return fmt.Errorf("unsupported expect value type: %T", expectVal)
			}
		}
	}
	return nil
}

// checkExpectations checks the expectations in the provided list.
func checkExpectations(expectations []interface{}, client *http.Client) error {
	strs := make([]string, len(expectations))
	for i, v := range expectations {
		if s, ok := v.(string); ok {
			strs[i] = s
		} else {
			return fmt.Errorf("unsupported expect value: %v", v)
		}
	}
	return expect.CheckExpectations("", 0, strs, client)
}

// isValidCheckPrefix checks if the string has a valid prefix for checks.
func isValidCheckPrefix(s string) bool {
	prefixes := []string{"ENV:", "FILE:", "DIR:", "URL:", "!"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")
}

// executeAndLogCommand executes the command for a step and logs the result.
func executeAndLogCommand(step RunStep, resName, resNode string, logs *logs, client *http.Client) error {
	LogInfo(fmt.Sprintf("Executing command: %s for resource: %s, step: %s", step.Exec, resName, step.Name))
	execResultChan := exec.ExecuteCommand(step.Exec)

	result, ok := <-execResultChan
	if !ok {
		return fmt.Errorf("failed to execute command: %s", step.Exec)
	}
	logEntry := stepLog{
		targetRes: resNode,
		command:   step.Exec,
		res:       resName,
		name:      step.Name,
		message:   result.Output,
	}
	logs.add(logEntry)

	if result.Err != nil {
		LogError(fmt.Sprintf("Command execution error for %s: %v", step.Name, result.Err), result.Err)
		return result.Err
	}

	if step.Expect != nil {
		expectations := expect.ProcessExpectations(step.Expect)
		if err := expect.CheckExpectations(result.Output, result.ExitCode, expectations, client); err != nil {
			LogError(fmt.Sprintf("Expectation check failed for %s: %v", step.Name, err), err)
			return err
		}
	}

	return nil
}

// HandleRunCommand handles the 'run' command for the given resources.
func (dr *DependencyResolver) HandleRunCommand(resources []string) error {
	logs := &logs{}

	visited := make(map[string]bool)
	client := &http.Client{}

	for _, resName := range resources {
		stack := dr.Graph.BuildDependencyStack(resName, visited)
		for _, resNode := range stack {
			for _, res := range dr.Resources {
				if res.Resource == resNode {
					dr.resolveDependency(resNode, res, logs, client)
				}
			}
		}
	}

	// Close the log after all processing is done.
	logs.close()

	return nil
}

// resolveDependency resolves the dependency for a given resource node.
func (dr *DependencyResolver) resolveDependency(resNode string, res ResourceEntry, logs *logs, client *http.Client) {
	LogInfo("🔍 Resolving dependency " + resNode)
	if res.Run == nil {
		return
	}

	skipResults := make(map[StepKey]bool)
	mu := &sync.Mutex{}

	for _, step := range res.Run {
		dr.processSkipSteps(step, resNode, skipResults, mu, client)
	}

	skip := dr.buildSkipMap(res.Run, resNode, skipResults)

	for _, step := range res.Run {
		dr.handleStep(step, resNode, skip, logs, client)
	}
}

// processSkipSteps processes skip steps for a given step.
func (dr *DependencyResolver) processSkipSteps(step RunStep, resNode string, skipResults map[StepKey]bool, mu *sync.Mutex, client *http.Client) {
	if skipSteps, ok := step.Skip.([]interface{}); ok {
		err := dr.processSteps(skipSteps, "skip", step.Name, client)
		dr.recordSkipResult(step, resNode, err == nil, skipResults, mu)
	} else {
		dr.recordSkipResult(step, resNode, false, skipResults, mu)
	}
}

// recordSkipResult records the result of a skip step.
func (dr *DependencyResolver) recordSkipResult(step RunStep, resNode string, result bool, skipResults map[StepKey]bool, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()
	skipResults[StepKey{name: step.Name, node: resNode}] = result
}

// buildSkipMap builds a map of skip results.
func (dr *DependencyResolver) buildSkipMap(steps []RunStep, resNode string, skipResults map[StepKey]bool) map[StepKey]bool {
	skip := make(map[StepKey]bool)
	for _, step := range steps {
		skip[StepKey{name: step.Name, node: resNode}] = skipResults[StepKey{name: step.Name, node: resNode}]
	}
	return skip
}

// handleStep handles the execution and logging of a step.
func (dr *DependencyResolver) handleStep(step RunStep, resNode string, skip map[StepKey]bool, logs *logs, client *http.Client) {
	skipKey := StepKey{name: step.Name, node: resNode}
	LogInfo(fmt.Sprintf("Step: %s, Skip: %v", step.Name, skip[skipKey]))

	if !skip[skipKey] {
		if checkSteps, ok := step.Check.([]interface{}); ok {
			if err := dr.processSteps(checkSteps, "check", step.Name, client); err != nil {
				LogError("Check expectation failed for resource '"+resNode+"' step '"+step.Name+"'", err)
				return
			}
		}

		// Execute the command and log the result.
		err := executeAndLogCommand(step, resNode, resNode, logs, client)
		if err != nil {
			LogError("Error executing command for resource '"+resNode+"' step '"+step.Name+"'", err)
			return
		}

		if step.Expect != nil {
			expectations := expect.ProcessExpectations(step.Expect)
			result, ok := <-exec.ExecuteCommand(step.Exec)
			if !ok {
				LogError("Failed to execute command: "+step.Exec, nil)
				return
			}
			if err := expect.CheckExpectations(result.Output, result.ExitCode, expectations, client); err != nil {
				LogError("Expectation check failed for resource '"+resNode+"' step '"+step.Name+"'", err)
			}
		}
	}
}

// HandleShowCommand handles the 'show' command for the given resources.
func (dr *DependencyResolver) HandleShowCommand(resources []string) error {
	for _, res := range resources {
		if err := dr.ShowResourceEntry(res); err != nil {
			return LogError("Error showing resource entry "+res, err)
		}
	}
	return nil
}

// HandleDependsCommand handles the 'depends' command for the given resources.
func (dr *DependencyResolver) HandleDependsCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListDirectDependencies(res)
	}
	return nil
}

// HandleRDependsCommand handles the 'rdepends' command for the given resources.
func (dr *DependencyResolver) HandleRDependsCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListReverseDependencies(res)
	}
	return nil
}

// HandleSearchCommand handles the 'search' command.
func (dr *DependencyResolver) HandleSearchCommand(resources []string) error {
	query := resources[0]
	keys := resources[1:]
	return dr.FuzzySearch(query, keys)
}

// HandleCategoryCommand handles the 'category' command for the given categories.
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

// HandleTreeCommand handles the 'tree' command for the given resources.
func (dr *DependencyResolver) HandleTreeCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListDependencyTree(res)
	}
	return nil
}

// HandleTreeListCommand handles the 'tree-list' command for the given resources.
func (dr *DependencyResolver) HandleTreeListCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListDependencyTreeTopDown(res)
	}
	return nil
}

// HandleIndexCommand handles the 'index' command, listing all resources.
func (dr *DependencyResolver) HandleIndexCommand() error {
	for _, entry := range dr.Resources {
		PrintMessage("📦 Resource: %s\n📛 Name: %s\n📝 Short Description: %s\n📖 Long Description: %s\n🏷️  Category: %s\n🔗 Requirements: %v\n",
			entry.Resource, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
		Println("---")
	}
	return nil
}
