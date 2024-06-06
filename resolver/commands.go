package resolver

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/kdeps/plugins/expect"
	"github.com/kdeps/plugins/kdepexec"
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
	fmt.Println(formatLogEntry(entry))
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

// getAllMessages retrieves all log messages as a slice of strings.
func (m *logs) getAllMessages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	messages := make([]string, len(m.entries))
	for i, entry := range m.entries {
		messages[i] = entry.message
	}
	return messages
}

// getAllMessageString retrieves all log messages as a string.
func (m *logs) getAllMessageString() string {
	return strings.Join(m.getAllMessages(), "\n")
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

func SourceEnvFile(envFilePath string) error {
	file, err := os.Open(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to open env file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid environment variable declaration: %s", line)
		}
		key := parts[0]
		value := strings.Trim(parts[1], "\"")
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %v", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %v", err)
	}

	return nil
}

// processSteps processes each step by executing the relevant checks.
func (dr *DependencyResolver) processSteps(steps []interface{}, stepType, resNode string, client *http.Client, logs *logs) error {
	for _, step := range steps {
		LogDebug(fmt.Sprintf("Processing '%s' step: '%v' - '%s'", stepType, step, resNode))
		if err := processElement(step, client, logs); err != nil {
			LogError(fmt.Sprintf("Error processing step '%v' in '%s' steps: ", step, stepType), err)
			return err
		}
	}
	return nil
}

// processElement processes an individual step element based on its type.
func processElement(element interface{}, client *http.Client, logs *logs) error {
	switch val := element.(type) {
	case string:
		if isValidCheckPrefix(val) {
			return expect.CheckExpectations(logs.getAllMessageString(), 0, []string{val}, client)
		} else {
			LogDebug(fmt.Sprintf("Skipping check condition '%s' unsupported.", val))
		}
	case map[interface{}]interface{}:
		if expectVal, exists := val["expect"]; exists {
			ev := expectVal.([]interface{})
			return checkExpectationsArray(ev, client, logs)
		}
	default:
		LogErrorExit(fmt.Sprintf("Unsupported Step: %v", val), nil)
	}
	return nil
}

// CheckExpectationsArray checks the expectations in the provided list.
func checkExpectationsArray(expectations []interface{}, client *http.Client, logs *logs) error {
	strs := make([]string, len(expectations))
	for i, v := range expectations {
		if s, ok := v.(string); ok {
			strs[i] = s
		}
	}
	return expect.CheckExpectations(logs.getAllMessageString(), 0, strs, client)
}

// isValidCheckPrefix checks if the string has a valid prefix for checks.
func isValidCheckPrefix(s string) bool {
	prefixes := []string{"ENV:", "FILE:", "DIR:", "URL:", "CMD:", "!"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")
}

func (dr *DependencyResolver) SetEnvironmentVariables(envVars []EnvVar) error {
	for _, envVar := range envVars {
		var value string
		if envVar.Exec != "" {
			var result kdepexec.CommandResult
			var ok bool

			resultChan := dr.ShellSession.ExecuteCommand(envVar.Exec)
			result, ok = <-resultChan

			if !ok {
				LogErrorExit(fmt.Sprintf("Failed to set ENV VAR: '%s'", envVar.Exec), nil)
			}
			value = result.Output
		} else {
			value = envVar.Value
		}
		if err := os.Setenv(envVar.Name, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %v", envVar.Name, err)
		}
	}
	return nil
}

func (dr *DependencyResolver) ExecuteAndLogCommand(step RunStep, resName, resNode string, logs *logs, client *http.Client) error {
	LogInfo(fmt.Sprintf("Executing command: '%s' for resource: '%s', step: '%s'", step.Exec, resName, step.Name))

	// Set environment variables
	if err := dr.SetEnvironmentVariables(step.Env); err != nil {
		LogErrorExit(fmt.Sprintf("Failed to set environment variables for step: '%s'", step.Name), err)
	}

	var result kdepexec.CommandResult
	var ok bool

	execResultChan := dr.ShellSession.ExecuteCommand(step.Exec)
	result, ok = <-execResultChan
	logEntry := stepLog{
		targetRes: resNode,
		command:   step.Exec,
		res:       resName,
		name:      step.Name,
		message:   result.Output,
	}
	logs.add(logEntry)

	if !ok {
		LogErrorExit(fmt.Sprintf("Failed to execute command: '%s'", step.Exec), nil)
	}

	if result.Err != nil {
		LogErrorExit(fmt.Sprintf("Command execution error for '%s' ", step.Name), result.Err)
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
		LogInfo("No run steps found for resource " + resNode)
		return
	}

	skipResults := make(map[StepKey]bool)
	mu := &sync.Mutex{}

	for _, step := range res.Run {
		dr.processSkipSteps(step, resNode, skipResults, mu, client, logs)
	}

	skip := dr.buildSkipMap(res.Run, resNode, skipResults)

	for _, step := range res.Run {
		dr.handleStep(step, resNode, skip, logs, client)
	}
}

// processSkipSteps processes skip steps for a given step.
func (dr *DependencyResolver) processSkipSteps(step RunStep, resNode string, skipResults map[StepKey]bool, mu *sync.Mutex, client *http.Client, logs *logs) {
	if skipSteps, ok := step.Skip.([]interface{}); ok {
		for _, skipStep := range skipSteps {
			if skipStr, ok := skipStep.(string); ok && isValidCheckPrefix(skipStr) {
				if err := processElement(skipStr, client, logs); err == nil {
					mu.Lock()
					skipResults[StepKey{name: step.Name, node: resNode}] = true
					mu.Unlock()

					LogDebug(fmt.Sprintf("Skipping step '%s' for node '%s' due to skip condition", step.Name, resNode))

					return
				}
			}
		}
		mu.Lock()
		skipResults[StepKey{name: step.Name, node: resNode}] = false
		mu.Unlock()

		LogDebug(fmt.Sprintf("Not skipping step '%s' for node '%s'", step.Name, resNode))
	}
}

// buildSkipMap builds a map of skip results.
func (dr *DependencyResolver) buildSkipMap(steps []RunStep, resNode string, skipResults map[StepKey]bool) map[StepKey]bool {
	skip := make(map[StepKey]bool)
	for _, step := range steps {
		skip[StepKey{name: step.Name, node: resNode}] = skipResults[StepKey{name: step.Name, node: resNode}]
	}
	return skip
}

// handleStep handles the kdepexecution and logging of a step.
func (dr *DependencyResolver) handleStep(step RunStep, resNode string, skip map[StepKey]bool, logs *logs, client *http.Client) {
	skipKey := StepKey{name: step.Name, node: resNode}
	LogDebug(fmt.Sprintf("Skip key '%v' = %v", skipKey, skip[skipKey]))

	if skip[skipKey] {
		logs.add(stepLog{targetRes: resNode, command: step.Exec, res: resNode, name: step.Name, message: "Step skipped."})
		LogInfo("Step: '" + step.Name + "' skipped for resource: '" + resNode + "'")
		return
	}

	if step.Exec != "" {
		if err := dr.ExecuteAndLogCommand(step, resNode, resNode, logs, client); err != nil {
			LogErrorExit(fmt.Sprintf("Execution failed for step '%s' of resource '%s': ", step.Name, resNode), err)
		}
	}

	if checkSteps, ok := step.Check.([]interface{}); ok {
		if err := dr.processSteps(checkSteps, "check", resNode, client, logs); err != nil {
			LogErrorExit("Check expectation failed for resource '"+resNode+"' step '"+step.Name+"'", err)
		}
	}

	if expectSteps, ok := step.Expect.([]interface{}); ok {
		if err := SourceEnvFile(os.Getenv("KDEPS_ENV")); err != nil {
			LogErrorExit(fmt.Sprintf("Failed to source environment file for step: '%s'", step.Name), err)
		}

		expectations := expect.ProcessExpectations(expectSteps)
		if err := expect.CheckExpectations(logs.getAllMessageString(), 0, expectations, client); err != nil {
			LogErrorExit(fmt.Sprintf("Expectation failed for '%s': ", step.Name), err)
		}
	}
}

// HandleShowCommand handles the 'show' command for the given resources.
func (dr *DependencyResolver) HandleShowCommand(resources []string) error {
	for _, res := range resources {
		if err := dr.ShowResourceEntry(res); err != nil {
			LogErrorExit("Error showing resource entry "+res, err)
		}
	}
	return nil
}

// HandleDependsCommand handles the 'depends' command for the given resources.
func (dr *DependencyResolver) HandleDependsCommand(resources []string) error {
	for _, res := range resources {
		LogDebug("Listing direct dependencies for resource " + res)
		dr.Graph.ListDirectDependencies(res)
	}
	return nil
}

// HandleRDependsCommand handles the 'rdepends' command for the given resources.
func (dr *DependencyResolver) HandleRDependsCommand(resources []string) error {
	for _, res := range resources {
		LogDebug("Listing reverse dependencies for resource " + res)
		dr.Graph.ListReverseDependencies(res)
	}
	return nil
}

// HandleSearchCommand handles the 'search' command.
func (dr *DependencyResolver) HandleSearchCommand(resources []string) error {
	query := resources[0]
	keys := resources[1:]
	LogDebug("Performing fuzzy search with query: " + query)
	return dr.FuzzySearch(query, keys)
}

// HandleCategoryCommand handles the 'category' command for the given categories.
func (dr *DependencyResolver) HandleCategoryCommand(resources []string) error {
	if len(resources) == 0 {
		LogInfo("No categories provided")
		Println("Usage: kdeps category [categories...]")
		return nil
	}
	for _, entry := range dr.Resources {
		for _, category := range resources {
			if entry.Category == category {
				LogDebug("Listing resource in category: " + category)
				Println("📂 " + entry.Resource)
			}
		}
	}
	return nil
}

// HandleTreeCommand handles the 'tree' command for the given resources.
func (dr *DependencyResolver) HandleTreeCommand(resources []string) error {
	for _, res := range resources {
		LogDebug("Listing dependency tree for resource " + res)
		dr.Graph.ListDependencyTree(res)
	}
	return nil
}

// HandleTreeListCommand handles the 'tree-list' command for the given resources.
func (dr *DependencyResolver) HandleTreeListCommand(resources []string) error {
	for _, res := range resources {
		LogDebug("Listing top-down dependency tree for resource " + res)
		dr.Graph.ListDependencyTreeTopDown(res)
	}
	return nil
}

// HandleIndexCommand handles the 'index' command, listing all resources.
func (dr *DependencyResolver) HandleIndexCommand() error {
	for _, entry := range dr.Resources {
		LogDebug("Indexing resource: " + entry.Resource)
		PrintMessage("📦 Resource: %s\n📛 Name: %s\n📝 Short Description: %s\n📖 Long Description: %s\n🏷️  Category: %s\n🔗 Requirements: %v\n",
			entry.Resource, entry.Name, entry.Sdesc, entry.Ldesc, entry.Category, entry.Requires)
		Println("---")
	}
	return nil
}
