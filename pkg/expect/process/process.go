package process

import (
	"os"
	"strconv"
	"strings"
)

// ReplaceVars replaces placeholders with environment variable values.
func ReplaceVars(expectation string) string {
	// This pattern matches ${VAR_NAME} and replaces it with the environment variable value.
	for {
		start := strings.Index(expectation, "${")
		if start == -1 {
			break
		}
		end := strings.Index(expectation[start:], "}")
		if end == -1 {
			break
		}
		end += start
		varName := expectation[start+2 : end]
		varValue := os.Getenv(varName)
		expectation = expectation[:start] + varValue + expectation[end+1:]
	}
	return expectation
}

// ProcessExpectations converts the expectations into a string slice.
func ProcessExpectations(expect interface{}) []string {
	var expectations []string

	switch v := expect.(type) {
	case string:
		expectations = []string{ReplaceVars(v)}
	case int:
		expectations = []string{strconv.Itoa(v)}
	case []interface{}:
		for _, item := range v {
			switch item := item.(type) {
			case string:
				expectations = append(expectations, ReplaceVars(item))
			case int:
				expectations = append(expectations, strconv.Itoa(item))
			}
		}
	}

	return expectations
}
