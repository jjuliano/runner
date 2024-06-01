package resolver

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kdeps/plugins/exec"
	"github.com/kdeps/plugins/expect"
)

type stepLog struct {
	name      string
	message   string
	res       string
	command   string
	targetRes string
}

type logs []stepLog

func (m *logs) addLogs(entry stepLog, logChan chan<- stepLog) {
	*m = append(*m, entry)
	logChan <- entry
}

func (m logs) getLogs() []stepLog {
	logEntries := make([]stepLog, len(m))
	copy(logEntries, m)
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

func (dr *DependencyResolver) processSteps(steps []interface{}, stepType, resNode string, client *http.Client) error {
	for _, step := range steps {
		LogInfo(fmt.Sprintf("Processing '%s' step: '%v' - '%s'", stepType, step, resNode))
		if err := processElement(step, client); err != nil {
			return err
		}
	}
	return nil
}

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

func isValidCheckPrefix(s string) bool {
	prefixes := []string{"ENV:", "FILE:", "DIR:", "URL:", "!"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")
}

func executeAndLogCommand(step RunStep, resName, resNode string, logs *logs, logChan chan<- stepLog, client *http.Client) error {
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
	logs.addLogs(logEntry, logChan)

	LogInfo(fmt.Sprintf("Command executed. Result: %v", result))

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
	defer close(logChan)

	for _, resName := range resources {
		stack := dr.Graph.BuildDependencyStack(resName, visited)
		for _, resNode := range stack {
			for _, res := range dr.Resources {
				if res.Resource == resNode {
					dr.resolveDependency(resNode, res, logs, logChan, client)
				}
			}
		}
	}
	return nil
}

func (dr *DependencyResolver) resolveDependency(resNode string, res ResourceEntry, logs *logs, logChan chan<- stepLog, client *http.Client) {
	LogInfo("🔍 Resolving dependency " + resNode)
	if res.Run == nil {
		return
	}

	var wg sync.WaitGroup
	skipResults := make(map[StepKey]bool)
	mu := &sync.Mutex{}

	for _, step := range res.Run {
		wg.Add(1)
		go dr.processSkipSteps(step, resNode, skipResults, mu, &wg, client)
	}
	wg.Wait()

	skip := dr.buildSkipMap(res.Run, resNode, skipResults)

	for _, step := range res.Run {
		dr.handleStep(step, resNode, skip, logs, logChan, client)
	}
}

func (dr *DependencyResolver) processSkipSteps(step RunStep, resNode string, skipResults map[StepKey]bool, mu *sync.Mutex, wg *sync.WaitGroup, client *http.Client) {
	defer wg.Done()
	if skipSteps, ok := step.Skip.([]interface{}); ok {
		err := dr.processSteps(skipSteps, "skip", step.Name, client)
		dr.recordSkipResult(step, resNode, err == nil, skipResults, mu)
	} else {
		dr.recordSkipResult(step, resNode, false, skipResults, mu)
	}
}

func (dr *DependencyResolver) recordSkipResult(step RunStep, resNode string, result bool, skipResults map[StepKey]bool, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()
	skipResults[StepKey{name: step.Name, node: resNode}] = result
}

func (dr *DependencyResolver) buildSkipMap(steps []RunStep, resNode string, skipResults map[StepKey]bool) map[StepKey]bool {
	skip := make(map[StepKey]bool)
	for _, step := range steps {
		skip[StepKey{name: step.Name, node: resNode}] = skipResults[StepKey{name: step.Name, node: resNode}]
	}
	return skip
}

func (dr *DependencyResolver) handleStep(step RunStep, resNode string, skip map[StepKey]bool, logs *logs, logChan chan<- stepLog, client *http.Client) {
	skipKey := StepKey{name: step.Name, node: resNode}
	LogInfo(fmt.Sprintf("Step: %s, Skip: %v", step.Name, skip[skipKey]))

	if !skip[skipKey] {
		if checkSteps, ok := step.Check.([]interface{}); ok {
			if err := dr.processSteps(checkSteps, "check", step.Name, client); err != nil {
				LogError("Check expectation failed for resource '"+resNode+"' step '"+step.Name+"'", err)
			}
		}

		if step.Exec != "" {
			LogInfo(fmt.Sprintf("Executing command for resource: %s, step: %s", resNode, step.Name))
			if err := executeAndLogCommand(step, resNode, resNode, logs, logChan, client); err != nil {
				LogError("Error executing command '"+resNode+"' step '"+step.Name+"'", err)
			}
		}
	}
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
		dr.Graph.ListDirectDependencies(res)
	}
	return nil
}

func (dr *DependencyResolver) HandleRDependsCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListReverseDependencies(res)
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
		dr.Graph.ListDependencyTree(res)
	}
	return nil
}

func (dr *DependencyResolver) HandleTreeListCommand(resources []string) error {
	for _, res := range resources {
		dr.Graph.ListDependencyTreeTopDown(res)
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
