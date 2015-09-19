package validate

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"github.com/codegangsta/cli"
	"github.com/marcosnils/cmt/cmd"
)

var Command = cli.Command{
	Name:  "validate",
	Usage: "Validate host migration capabilities",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "src",
			Usage: "Source host URL [user@host:port]",
		},
		cli.StringFlag{
			Name:  "dst",
			Usage: "Destination host URL [user@host:port]",
		},
	},
	Action: func(c *cli.Context) {
		src := c.String("src")
		if src != "" && !strings.HasPrefix(src, "ssh://") {
			src = fmt.Sprintf("ssh://%s", src)
		}

		dst := c.String("dst")
		if dst != "" && !strings.HasPrefix(dst, "ssh://") {
			dst = fmt.Sprintf("ssh://%s", dst)
		}

		srcURL := parseURL(src)
		dstURL := parseURL(dst)

		Validate(srcURL, dstURL)
		println("Validation succeded")

	},
}

func parseURL(stringURL string) *url.URL {
	if stringURL == "" {
		return nil
	}

	parsedURL, err := url.Parse(stringURL)
	if err != nil || parsedURL.Host == "" {
		log.Fatal("Error parsing host: ", stringURL)
	}

	return parsedURL

}

func Validate(src, dst *url.URL) {
	if src == nil && dst == nil {
		log.Fatal("Either one of dst or src must be specified")
	}

	srcCmd := getCommand(src)
	dstCmd := getCommand(dst)

	if e := checkVersion(srcCmd, dstCmd, "criu"); e != nil {
		log.Fatal(e)
	}
	if e := checkVersion(srcCmd, dstCmd, "runc"); e != nil {
		log.Fatal(e)
	}

	if e := checkKernelCap(srcCmd); e != nil {
		log.Fatal(e)
	}

	if e := checkKernelCap(dstCmd); e != nil {
		log.Fatal(e)
	}

	if e := checkCPUCompat(srcCmd, dstCmd); e != nil {
		log.Fatal(e)
	}
}

func checkCPUCompat(srcCmd, dstCmd cmd.Cmd) error {
	// Dump
	_, _, err := srcCmd.Run("sudo", "criu", "cpuinfo", "dump")
	if _, ok := err.(*ssh.ExitError); ok {
		return fmt.Errorf("Error dumping CPU info")
	} else if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("Error dumping CPU info")
	} else if err != nil {
		return fmt.Errorf("Connection error: %s ", err)
	}

	// Copy

	err = cmd.Scp(srcCmd.URL("./cpuinfo.img"), dstCmd.URL("."))
	if _, ok := err.(*ssh.ExitError); ok {
		return fmt.Errorf("Error copying dump image")
	} else if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("Error copying dump image")
	} else if err != nil {
		return fmt.Errorf("Connection error: %s ", err)
	}

	// Check
	_, _, err = srcCmd.Run("sudo", "criu", "cpuinfo", "check")
	if _, ok := err.(*ssh.ExitError); ok {
		return fmt.Errorf("Error checking CPU info")
	} else if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("Error checking CPU info")
	} else if err != nil {
		return fmt.Errorf("Connection error: %s ", err)
	}
	return nil
}

func checkKernelCap(c cmd.Cmd) error {
	_, _, err := c.Run("sudo", "criu", "check", "--ms")
	if _, ok := err.(*ssh.ExitError); ok {
		return fmt.Errorf("Error criu checks do not pass")
	} else if _, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("Error criu checks do not pass")
	} else if err != nil {
		return fmt.Errorf("Connection error: %s ", err)
	}
	return err
}

func getCommand(hostURL *url.URL) cmd.Cmd {
	if hostURL != nil {
		rc := cmd.NewSSH(hostURL.User.Username(), hostURL.Host)
		if err := rc.UseAgent(); err != nil {
			log.Fatal("Unable to use SSH agent for host: ", hostURL.String())
		}
		return rc

	}

	return cmd.NewLocal()

}

func checkVersion(sCmd, dCmd cmd.Cmd, name string) error {
	var wg sync.WaitGroup
	wg.Add(2)
	var sourceVersion, destVersion string
	var sourceError, destError error
	go func() {
		sourceVersion, sourceError = getVersion(sCmd, name)
		wg.Done()
	}()
	go func() {
		destVersion, destError = getVersion(dCmd, name)
		wg.Done()
	}()

	wg.Wait()

	if sourceError != nil {
		return fmt.Errorf("%s in src", sourceError)
	}
	if destError != nil {
		return fmt.Errorf("%s in dst", destError)
	}

	if sourceVersion != destVersion {
		return fmt.Errorf("ERROR: Source and destination versions of %s do not match", name)
	}

	return nil
}

func getVersion(command cmd.Cmd, name string) (string, error) {
	version, _, err := command.Run("sudo", name, "--version")
	if _, ok := err.(*ssh.ExitError); ok {
		return "", fmt.Errorf("Error %s does not exist", name)
	} else if _, ok := err.(*exec.ExitError); ok {
		return "", fmt.Errorf("Error %s does not exist", name)
	} else if err != nil {
		return "", fmt.Errorf("Connection error: %s ", err)
	}
	return version, nil
}
