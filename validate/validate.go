package validate

import "github.com/codegangsta/cli"

var Command = cli.Command{
	Name:  "validate",
	Usage: "Validate host migration capabilities",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Usage: "Host to run migration validation against",
		},
	},
	Action: func(c *cli.Context) {
		println("validate")
	},
}
