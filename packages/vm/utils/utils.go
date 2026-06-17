package utils

import (
	"bytes"
	"os"
	"os/exec"
)

func Run(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}

func RunWithOutput(cmd string, args ...string) (string, string, error) {
	c := exec.Command(cmd, args...)
	var output bytes.Buffer
	var errOutput bytes.Buffer

	c.Stdout = &output
	c.Stderr = &errOutput

	return output.String(), errOutput.String(), c.Run()
}
