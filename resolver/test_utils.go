package resolver

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// captureOutput captures the standard output for testing
func captureOutput(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Println("Error creating pipe:", err)
		return ""
	}

	old := os.Stdout
	os.Stdout = w

	outC := make(chan string)
	// Copy the output in a separate goroutine so that it doesn't block.
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			fmt.Println("Error copying output:", err)
		}
		outC <- buf.String()
	}()

	// Execute the function
	f()

	// Restore the original stdout and close the pipe
	err = w.Close()
	if err != nil {
		fmt.Println("Error closing pipe:", err)
	}
	os.Stdout = old
	return <-outC
}

// getSecondStrings is a utility function used by FuzzySearch
func getSecondStrings(entries [][2]string) []string {
	strs := make([]string, len(entries))
	for i, entry := range entries {
		strs[i] = entry[1]
	}
	return strs
}
