package main

import (
	"log"

	"github.com/marcosnils/cmt/cmd"
)

func main() {
	c := cmd.NewSSH("ubuntu", "172.31.60.136", 22)
	err := c.UsePrivateKey("/home/ubuntu/.ssh/docker_global.pem")
	err := c.UseAgent()

	// c := cmd.NewLocal()

	if err != nil {
		log.Fatal(err)
	}
	stdout, stderr, err := c.Run("ls", "-la")

	log.Println(stdout, stderr, err)
}
