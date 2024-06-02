package resolver

import (
	"bytes"
	"io"
	"os"
)

// captureOutput captures the standard output for testing
func captureOutput(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		PrintError("Error creating pipe", err)
		return ""
	}

	old := os.Stdout
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			PrintError("Error copying output", err)
		}
		outC <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = old
	return <-outC
}
