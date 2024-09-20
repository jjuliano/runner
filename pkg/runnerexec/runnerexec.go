package runnerexec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CommandResult holds the output, exit code, and error of a command execution.
type CommandResult struct {
	Output   string
	ExitCode int
	Err      error
}

type ShellSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

// NewShellSession creates a new shell session
func NewShellSession() (*ShellSession, error) {
	cmd := exec.Command("sh")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &ShellSession{cmd: cmd, stdin: stdin, stdout: stdout, stderr: stderr}, nil
}

// RunCommand sends a command to the shell session
func (s *ShellSession) RunCommand(command string) error {
	_, err := s.stdin.Write([]byte(command + "\n"))
	return err
}

// Close terminates the shell session
func (s *ShellSession) Close() error {
	s.stdin.Close()
	return s.cmd.Wait()
}

// ExecuteCommand runs a shell command and returns its output, exit code, and error if any.
func (s *ShellSession) ExecuteCommand(execCmd string) <-chan CommandResult {
	resultChan := make(chan CommandResult)

	go func() {
		defer close(resultChan)

		var outbuf, errbuf bytes.Buffer

		// Use a new command to execute the input command within the session
		cmd := exec.Command("sh", "-c", execCmd)
		cmd.Stdout = &outbuf
		cmd.Stderr = &errbuf

		err := cmd.Run()
		output := outbuf.String() + errbuf.String()
		exitCode := 0

		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			}
		}

		resultChan <- CommandResult{Output: output, ExitCode: exitCode, Err: err}
	}()

	return resultChan
}

// Which searches for an executable in the directories specified by the PATH environment variable.
func Which(executable string) (string, error) {
	pathEnv := os.Getenv("PATH")
	pathSeparator := string(os.PathListSeparator)
	paths := strings.Split(pathEnv, pathSeparator)

	var extensions []string
	if runtime.GOOS == "windows" {
		pathext := os.Getenv("PATHEXT")
		extensions = strings.Split(pathext, pathSeparator)
		extensions = append(extensions, "")
	} else {
		extensions = []string{""}
	}

	for _, dir := range paths {
		for _, ext := range extensions {
			fullPath := filepath.Join(dir, executable+ext)
			if fileInfo, err := os.Stat(fullPath); err == nil {
				mode := fileInfo.Mode()
				if mode.IsRegular() && (mode&0111 != 0) {
					return fullPath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("%s: command not found", executable)
}
