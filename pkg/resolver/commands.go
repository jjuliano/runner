package resolver

import (
    "bufio"
    "fmt"
    "net/http"
    "os"
    "strings"
    "sync"

    "github.com/jjuliano/runner/pkg/expect"
    "github.com/jjuliano/runner/pkg/runnerexec"
)

// StepLog represents the structure of a log entry for a step.
type StepLog struct {
    name      string
    message   string
    id        string
    command   string
    targetRes string
}

// RunnerLogs manages the logging mechanism with synchronization.
type RunnerLogs struct {
    mu      sync.Mutex
    entries []StepLog
    closed  bool
}

// Add adds a new log entry to the RunnerLogs.
func (m *RunnerLogs) Add(entry StepLog) {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.closed {
        return
    }
    fmt.Println(FormatLogEntry(entry))
    m.entries = append(m.entries, entry)
}

// Close closes the log after all goroutines are done.
func (m *RunnerLogs) Close() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.closed = true
}

// StepLogs retrieves all log entries.
func (m *RunnerLogs) StepLogs() []StepLog {
    m.mu.Lock()
    defer m.mu.Unlock()
    logEntries := make([]StepLog, len(m.entries))
    copy(logEntries, m.entries)
    return logEntries
}

// GetAllMessages retrieves all log messages as a slice of strings.
func (m *RunnerLogs) GetAllMessages() []string {
    m.mu.Lock()
    defer m.mu.Unlock()
    messages := make([]string, len(m.entries))
    for i, entry := range m.entries {
        messages[i] = entry.message
    }
    return messages
}

// GetAllMessageString retrieves all log messages as a string.
func (m *RunnerLogs) GetAllMessageString() string {
    return strings.Join(m.GetAllMessages(), "\n")
}

// FormatLogEntry formats a log entry into a string.
func FormatLogEntry(entry StepLog) string {
    return strings.Join([]string{
        "\n",
        "📦 Id: " + entry.id,
        "📛 Step: " + entry.name,
        "📝 Command: " + entry.command,
        "\n" + entry.message,
    }, "\n")
}

func SourceEnvFile(envFilePath string) error {
    LogDebug(fmt.Sprintf("Sourcing environment file from path: %s", envFilePath))

    file, err := os.Open(envFilePath)
    if err != nil {
        return LogError(fmt.Sprintf("Failed to open environment file: %s - %v", envFilePath, err), err)
    }
    defer func(file *os.File) {
        err := file.Close()
        if err != nil {
            LogError(fmt.Sprintf("Failed to close environment file: %s - %v", envFilePath, err), err)
        } else {
            LogInfo(fmt.Sprintf("Successfully closed environment file: %s", envFilePath))
        }
    }(file)

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        LogDebug(fmt.Sprintf("Processing line: %s", line))

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            return LogError(fmt.Sprintf("Invalid environment variable declaration: %s in file: %s", line, envFilePath), err)
        }

        key := parts[0]
        value := strings.Trim(parts[1], "\"")
        if err := os.Setenv(key, value); err != nil {
            return LogError(fmt.Sprintf("Failed to set environment variable %s: %v - %s", key, err, envFilePath), err)
        }
        LogInfo(fmt.Sprintf("Set environment variable %s=%s", key, value))
    }

    if err := scanner.Err(); err != nil {
        return LogError(fmt.Sprintf("Error reading environment file: %s - %v", envFilePath, err), err)
    }

    LogInfo(fmt.Sprintf("Successfully sourced environment file: %s", envFilePath))
    return nil
}

// ProcessNodeSteps processes each step by executing the relevant checks.
func (dr *DependencyResolver) ProcessNodeSteps(steps []interface{}, stepType, resNode string, client *http.Client, logs *RunnerLogs) error {
    for _, step := range steps {
        LogInfo(fmt.Sprintf("Processing '%s' step: '%v' - '%s'", stepType, step, resNode))
        if err := ProcessSingleNodeRule(step, client, logs); err != nil {
            return LogError(fmt.Sprintf("Error processing step '%v' in '%s' steps: ", step, stepType), err)
        }
    }
    return nil
}

// ProcessSingleNodeRule processes an individual step element based on its type.
func ProcessSingleNodeRule(element interface{}, client *http.Client, logs *RunnerLogs) error {
    switch val := element.(type) {
    case string:
        if HasValidRulePrefix(val) {
            return expect.CheckExpectations(logs.GetAllMessageString(), 0, []string{val}, client)
        } else {
            LogDebug(fmt.Sprintf("Skipping check condition '%s' unsupported.", val))
        }
    case map[interface{}]interface{}:
        if expectVal, exists := val["expect"]; exists {
            ev := expectVal.([]interface{})
            return ProcessResourceNodeRules(ev, client, logs)
        }
    default:
        LogErrorExit(fmt.Sprintf("Unsupported Step: %v", val), nil)
    }
    return nil
}

// ProcessResourceNodeRules checks the expectations in the provided list.
func ProcessResourceNodeRules(expectations []interface{}, client *http.Client, logs *RunnerLogs) error {
    strs := make([]string, len(expectations))
    for i, v := range expectations {
        if s, ok := v.(string); ok {
            strs[i] = s
        }
    }
    return expect.CheckExpectations(logs.GetAllMessageString(), 0, strs, client)
}

// HasValidRulePrefix checks if the string has a valid prefix for checks.
func HasValidRulePrefix(s string) bool {
    prefixes := []string{"ENV:", "FILE:", "DIR:", "URL:", "CMD:", "EXEC:", "!"}
    for _, prefix := range prefixes {
        if strings.HasPrefix(s, prefix) {
            return true
        }
    }
    return strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")
}

func (dr *DependencyResolver) ProcessResourceNodeEnvVarDeclarations(envVars []EnvVar) error {
    for _, envVar := range envVars {
        var value string

        if envVar.Exec != "" {
            var result runnerexec.CommandResult
            var ok bool

            resultChan := dr.ShellSession.ExecuteCommand(envVar.Exec)
            result, ok = <-resultChan

            if !ok {
                LogErrorExit(fmt.Sprintf("Failed to set ENV VAR: '%s'", envVar.Exec), nil)
            }
            value = result.Output
        } else if envVar.Input != "" {
            fmt.Print(envVar.Input + ": ")

            _, err := fmt.Scanln(&value)
            if err != nil {
                LogErrorExit(fmt.Sprintf("Failed to read input for environment variable %s: ", envVar.Name), err)
            }
        } else if envVar.File != "" {
            // Check if envVar.File starts with a "$" to resolve environment variable
            if strings.HasPrefix(envVar.File, "$") {
                envVarName := envVar.File[1:] // Remove the "$" prefix
                filePath := os.Getenv(envVarName)
                if filePath == "" {
                    LogErrorExit(fmt.Sprintf("Environment variable %s not set or empty", envVarName), nil)
                }
                envVar.File = filePath
            }
        } else {
            value = envVar.Value
        }

        if err := os.Setenv(envVar.Name, value); err != nil {
            LogErrorExit(fmt.Sprintf("Failed to set environment variable %s: ", envVar.Name), err)
        }
    }
    return nil
}

func (dr *DependencyResolver) ExecuteAndLogCommand(step RunStep, resName string, resNode string, logs *RunnerLogs) error {
    LogInfo(fmt.Sprintf("Executing command: '%s' for resource: '%s', step: '%s'", step.Exec, resName, step.Name))

    // Set environment variables
    if err := dr.ProcessResourceNodeEnvVarDeclarations(step.Env); err != nil {
        LogErrorExit(fmt.Sprintf("Failed to set environment variables for step: '%s'", step.Name), err)
    }

    var result runnerexec.CommandResult
    var ok bool

    execResultChan := dr.ShellSession.ExecuteCommand(step.Exec)
    result, ok = <-execResultChan
    logEntry := StepLog{
        targetRes: resNode,
        command:   step.Exec,
        id:        resName,
        name:      step.Name,
        message:   result.Output,
    }
    logs.Add(logEntry)

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
    logs := &RunnerLogs{}

    visited := make(map[string]bool)
    client := &http.Client{}

    for _, resName := range resources {
        stack := dr.Graph.BuildDependencyStack(resName, visited)
        for _, resNode := range stack {
            for _, res := range dr.Resources {
                if res.Id == resNode {
                    dr.ResolveResourceNodeDependency(resNode, res, logs, client)
                }
            }
        }
    }

    // Close the log after all processing is done.
    logs.Close()

    return nil
}

// ResolveResourceNodeDependency resolves the dependency for a given resource node.
func (dr *DependencyResolver) ResolveResourceNodeDependency(resNode string, res ResourceNodeEntry, logs *RunnerLogs, client *http.Client) {
    LogInfo("ð Resolving dependency " + resNode)
    if res.Run == nil {
        LogInfo("No run steps found for resource " + resNode)
        return
    }

    skipResults := make(map[StepKey]bool)
    mu := &sync.Mutex{}

    for _, step := range res.Run {
        dr.ProcessNodeSkipRules(step, resNode, skipResults, mu, client, logs)
    }

    skip := dr.BuildNodeSkipMap(res.Run, resNode, skipResults)

    for _, step := range res.Run {
        dr.HandleResourceNodeStep(step, resNode, skip, logs, client)
    }
}

// ProcessNodeSkipRules processes skip steps for a given step.
func (dr *DependencyResolver) ProcessNodeSkipRules(step RunStep, resNode string, skipResults map[StepKey]bool, mu *sync.Mutex, client *http.Client, logs *RunnerLogs) {
    if skipSteps, ok := step.Skip.([]interface{}); ok {
        for _, skipStep := range skipSteps {
            if skipStr, ok := skipStep.(string); ok && HasValidRulePrefix(skipStr) {
                if err := ProcessSingleNodeRule(skipStr, client, logs); err == nil {
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

// BuildNodeSkipMap builds a map of skip results.
func (dr *DependencyResolver) BuildNodeSkipMap(steps []RunStep, resNode string, skipResults map[StepKey]bool) map[StepKey]bool {
    skip := make(map[StepKey]bool)
    for _, step := range steps {
        skip[StepKey{name: step.Name, node: resNode}] = skipResults[StepKey{name: step.Name, node: resNode}]
    }
    return skip
}

// HandleResourceNodeStep handles the execution and logging of a step.
func (dr *DependencyResolver) HandleResourceNodeStep(step RunStep, resNode string, skip map[StepKey]bool, logs *RunnerLogs, client *http.Client) {
    skipKey := StepKey{name: step.Name, node: resNode}
    // LogDebug(fmt.Sprintf("Skip key '%v' = %v", skipKey, skip[skipKey]))

    if skip[skipKey] {
        logs.Add(StepLog{targetRes: resNode, command: step.Exec, id: resNode, name: step.Name, message: "Step skipped."})
        LogInfo("Step: '" + step.Name + "' skipped for resource: '" + resNode + "'")
        return
    }

    if step.Exec != "" {
        if err := dr.ExecuteAndLogCommand(step, resNode, resNode, logs); err != nil {
            LogErrorExit(fmt.Sprintf("Execution failed for step '%s' of resource '%s': ", step.Name, resNode), err)
        }
    }

    if checkSteps, ok := step.Check.([]interface{}); ok {
        if err := dr.ProcessNodeSteps(checkSteps, "check", resNode, client, logs); err != nil {
            LogErrorExit("Check expectation failed for resource '"+resNode+"' step '"+step.Name+"'", err)
        }
    }

    if expectSteps, ok := step.Expect.([]interface{}); ok {
        if err := SourceEnvFile(os.Getenv("RUNNER_ENV")); err != nil {
            LogErrorExit(fmt.Sprintf("Failed to source environment file for step: '%s'", step.Name), err)
        }

        expectations := expect.ProcessExpectations(expectSteps)
        if err := expect.CheckExpectations(logs.GetAllMessageString(), 0, expectations, client); err != nil {
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
        Println("Usage: runner category [categories...]")
        return nil
    }
    for _, entry := range dr.Resources {
        for _, category := range resources {
            if entry.Category == category {
                LogDebug("Listing resource in category: " + category)
                Println("📦 " + entry.Id)
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
        LogDebug("Indexing resource: " + entry.Id)
        PrintMessage("📦 Id: %s\n📛 Name: %s\n📝 Description: %s\n🏷️  Category: %s\n🔗 Requirements: %v\n",
            entry.Id, entry.Name, entry.Desc, entry.Category, entry.Requires)
        fmt.Println()
    }
    return nil
}
