package migrate

import "github.com/codegangsta/cli"

var Command = cli.Command{
	Name:  "migrate",
	Usage: "Migrate running container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "source-host",
			Usage: "Source host where the container is running",
		},
		cli.StringFlag{
			Name:  "container-dir",
			Usage: "Directory where the container is running in the source host",
		},
		cli.StringFlag{
			Name:  "target-host",
			Usage: "Target host to migrate the container",
		},
		cli.StringFlag{
			Name:  "target-container-dir",
			Usage: "Directory to copy container data in target host",
		},
		cli.BoolFlag{
			Name:  "pre-dump",
			Usage: "Perform pre-dumps when migrating",
		},
		cli.DurationFlag{
			Name:  "max-downtime",
			Usage: "Max downtime allowed when performing pre-dumps",
		},
	},
	Action: func(c *cli.Context) {
	},
}
