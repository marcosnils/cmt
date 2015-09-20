package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHCmd struct {
	config         ssh.ClientConfig
	client         *ssh.Client
	host           string
	connected      bool
	currentSession *ssh.Session
}

func NewSSH(user, host string) *SSHCmd {
	c := SSHCmd{}
	c.config.User = user
	chunks := strings.Split(host, ":")
	c.host = chunks[0]
	return &c
}

func (r *SSHCmd) UseAgent() error {
	if r.connected {
		return fmt.Errorf("Cannot add authentication methods while being connected")
	}
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return err
	}
	r.config.Auth = append(r.config.Auth, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))

	return nil
}

func (r *SSHCmd) UsePrivateKey(path string) error {
	if r.connected {
		return fmt.Errorf("Cannot add authentication methods while being connected")
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	privateKey, err := ssh.ParseRawPrivateKey(content)
	if err != nil {
		return err
	}
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return err
	}

	authMethod := ssh.PublicKeys(signer)
	r.config.Auth = append(r.config.Auth, authMethod)

	return nil
}

func (r *SSHCmd) connect() error {
	if r.connected {
		return nil
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", r.host, 22), &r.config)
	if err != nil {
		return err
	}

	r.client = client
	r.connected = true

	return nil
}

func (r *SSHCmd) Run(name string, args ...string) (string, string, error) {
	if !r.connected {
		err := r.connect()
		if err != nil {
			return "", "", err
		}
	}

	session, err := r.client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(fmt.Sprintf("%s %s", name, strings.Join(args, " ")))

	return stdout.String(), stderr.String(), err
}

func (r *SSHCmd) Start(name string, args ...string) (Cmd, error) {
	if !r.connected {
		err := r.connect()
		if err != nil {
			return nil, err
		}
	}

	session, err := r.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return r, session.Start(fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
}

func (r *SSHCmd) Wait() error {
	if r.currentSession == nil {
		return errors.New("Start needs to be called before wait")
	}

	defer func() {
		r.currentSession.Close()
		r.currentSession = nil
	}()

	return r.currentSession.Wait()
}

func (r *SSHCmd) Output(name string, args ...string) (string, string, error) {
	log.Println(name, args)
	stdout, stderr, err := r.Run(name, args...)
	log.Println(stdout, stderr)
	return stdout, stderr, err
}

func (c *SSHCmd) URL(path string) *url.URL {
	return &url.URL{
		User: url.User(c.config.User),
		Host: c.host,
		Path: path,
	}

}
