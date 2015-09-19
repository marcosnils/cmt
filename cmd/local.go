package cmd

import (
	"bytes"
	"log"
	"net/url"
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

func (c *LocalCmd) Output(name string, args ...string) (string, string, error) {
	log.Println(name, args)
	stdout, stderr, err := c.Run(name, args...)
	log.Println(stdout, stderr)
	return stdout, stderr, err
}

func (c *LocalCmd) URL(path string) *url.URL {
	return &url.URL{
		Path: path,
	}

}
