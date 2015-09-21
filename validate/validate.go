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
		srcURL := ParseURL(c.String("src"))
		dstURL := ParseURL(c.String("dst"))

		Validate(srcURL, dstURL, false)
		println("Validation succeded")

	},
}

func ParseURL(rawurl string) *url.URL {
	if rawurl == "" {
		return nil
	}
	// We do this hack beacuse url.Parse require a schema to do the right thing
	schemaUrl := rawurl
	if !strings.HasPrefix(rawurl, "ssh://") {
		schemaUrl = fmt.Sprintf("ssh://%s", rawurl)
	}

	u, err := url.Parse(schemaUrl)
	if err != nil {
		log.Fatal("Error parsing host: ", rawurl)
	}

	return u

}

func Validate(src, dst *url.URL, continueOnWarnings bool) (srcCmd, dstCmd cmd.Cmd) {
	if src == nil || dst == nil {
		log.Fatal("Both src and dst must be specified")
	}

	srcCmd = GetCommand(src)
	dstCmd = GetCommand(dst)

	if found, e := checkVersion(srcCmd, dstCmd, "criu"); e != nil {
		if found && continueOnWarnings {
			log.Println("Warning: ", e)
		} else {
			log.Fatal(e)
		}
	}
	if found, e := checkVersion(srcCmd, dstCmd, "runc"); e != nil {
		if found && continueOnWarnings {
			log.Println("Warning: ", e)
		} else {
			log.Fatal(e)
		}
	}

	if e := checkKernelCap(srcCmd); e != nil {
		log.Fatal(e)
	}

	if e := checkKernelCap(dstCmd); e != nil {
		log.Fatal(e)
	}

	if e := checkCPUCompat(srcCmd, dstCmd); e != nil {
		if continueOnWarnings {
			log.Println("Warning: ", e)
		} else {
			log.Fatal(e)
		}
	}

	return
}

func checkCPUCompat(srcCmd, dstCmd cmd.Cmd) error {
	// Dump
	_, _, err := srcCmd.Run("criu", "cpuinfo", "dump")
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
	_, _, err = srcCmd.Run("criu", "cpuinfo", "check")
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

func GetCommand(hostURL *url.URL) cmd.Cmd {
	if hostURL.Host != "" {
		rc := cmd.NewSSH(hostURL.User.Username(), hostURL.Host)
		if err := rc.UseAgent(); err != nil {
			log.Fatal("Unable to use SSH agent for host: ", hostURL.String())
		}
		return rc

	}

	return cmd.NewLocal()

}

func checkVersion(sCmd, dCmd cmd.Cmd, name string) (bool, error) {
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
		return false, fmt.Errorf("%s in src", sourceError)
	}
	if destError != nil {
		return false, fmt.Errorf("%s in dst", destError)
	}

	if sourceVersion != destVersion {
		return true, fmt.Errorf("ERROR: Source and destination versions of %s do not match", name)
	}

	return true, nil
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
