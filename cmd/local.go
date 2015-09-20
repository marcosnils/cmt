package cmd

import (
	"bytes"
	"errors"
	"log"
	"net/url"
	"os/exec"
)

type LocalCmd struct {
	currentCommand *exec.Cmd
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

func (c *LocalCmd) Start(name string, args ...string) (Cmd, error) {
	command := exec.Command(name, args...)
	c.currentCommand = command
	return c, command.Start()
}

func (c *LocalCmd) Wait() error {
	if c.currentCommand == nil {
		return errors.New("Start needs to be called before wait")

	}
	defer func() {
		// Clear out current command
		c.currentCommand = nil

	}()
	return c.currentCommand.Wait()
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
