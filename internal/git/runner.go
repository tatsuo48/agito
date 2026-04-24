package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Runner abstracts external command execution.
type Runner interface {
	Run(name string, args ...string) (string, error)
}

// RealRunner executes commands via os/exec.
type RealRunner struct{}

func (r *RealRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%w: %s %v: %s", err, name, args, msg)
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}

// FakeRunner is a test stub that returns configured output.
type FakeRunner struct {
	Output string
	Err    error
	Calls  [][]string
}

func (f *FakeRunner) Run(name string, args ...string) (string, error) {
	call := append([]string{name}, args...)
	f.Calls = append(f.Calls, call)
	return f.Output, f.Err
}
