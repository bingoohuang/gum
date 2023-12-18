package utils

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LipglossPadding calculates how much padding a string is given by a style.
func LipglossPadding(style lipgloss.Style) (int, int) {
	render := style.Render(" ")
	before := strings.Index(render, " ")
	after := len(render) - len(" ") - before
	return before, after
}

type Result struct {
	result string
}

func (o *Result) SetResult(result string) { o.result = result }
func (o Result) GetResult() string        { return o.result }

type GetResult interface {
	GetResult() string
}

const ShellToUse = "bash"

func Shellout(command string) (stdOut, stdErr string, err error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}
