package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/marcosnils/cmt/migrate"
	"github.com/marcosnils/cmt/validate"
)

const (
	version = "0.1"
	usage   = `Container Migration Tool

cmt is a Docker Global Hackday #3 project. 
The purpose of the project is to create an external command line tool 
that can be either used with docker or runC which helps on the task to live migrate 
containers between different hosts by performing pre-migration validations
and allowing to auto-discover suitable target hosts.`
)

func main() {
	app := cli.NewApp()
	app.Name = "cmt"
	app.Usage = usage
	app.Version = version
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		migrate.Command,
		validate.Command,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
