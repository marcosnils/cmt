package cmd

import (
	"bytes"
	"os/exec"
)

type LocalCmd struct {
}

func NewLocal() Cmd {
	return &LocalCmd{}
}

func (c *LocalCmd) Run(name string, args ...string) (string, string, error) {
	command := exec.Command(name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	return stdout.String(), stderr.String(), err
}
