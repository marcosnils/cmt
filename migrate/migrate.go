package migrate

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/marcosnils/cmt/cmd"
	"github.com/marcosnils/cmt/validate"
)

var Command = cli.Command{
	Name:  "migrate",
	Usage: "Migrate running container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "src",
			Usage: "Source host where the container is running",
		},
		cli.StringFlag{
			Name:  "dst",
			Usage: "Target host to migrate the container",
		},
	},
	Action: func(c *cli.Context) {
		srcUrl := validate.ParseURL(c.String("src"))
		dstUrl := validate.ParseURL(c.String("dst"))

		log.Println("Performing validations")
		src, dst := validate.Validate(srcUrl, dstUrl)

		log.Println("Preparing everything to do a checkpoint")
		imagesPath := fmt.Sprintf("%s/images", srcUrl.Path)
		containerId := getContainerId(srcUrl.Path)

		_, _, err := src.Run("mkdir", "-p", imagesPath)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Performing the checkpoint")
		_, _, err = src.Run("sudo", "runc", "--id", containerId, "checkpoint", "--image-path", imagesPath)
		if err != nil {
			log.Fatal(err)
		}

		srcTarFile := fmt.Sprintf("%s/dump.tar.gz", srcUrl.Path)
		dstTarFile := fmt.Sprintf("%s/images/dump.tar.gz", dstUrl.Path)
		_, _, err = src.Run("sudo", "tar", "-czf", srcTarFile, "-C", fmt.Sprintf("%s/", imagesPath), ".")
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Copying checkpoint image to dst")
		_, _, err = dst.Run("mkdir", "-p", fmt.Sprintf("%s/images", dstUrl.Path))
		if err != nil {
			log.Fatal(err)
		}

		err = cmd.Scp(src.URL(srcTarFile), dst.URL(fmt.Sprintf("%s/images", dstUrl.Path)))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Preparing image at destination host")
		_, _, err = dst.Run("sudo", "tar", "-C", fmt.Sprintf("%s/images", dstUrl.Path), "-xvzf", dstTarFile)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Performing the restore")
		configFilePath := fmt.Sprintf("%s/config.json", dstUrl.Path)
		runtimeFilePath := fmt.Sprintf("%s/runtime.json", dstUrl.Path)
		dstImagesPath := fmt.Sprintf("%s/images", dstUrl.Path)
		_, _, err = dst.Output("sudo", "runc", "--id", containerId, "restore", "--image-path", dstImagesPath, "--config-file", configFilePath, "--runtime-file", runtimeFilePath)
		if err != nil {
			log.Fatal(err)
		}

	},
}

func getContainerId(path string) string {
	_, id := filepath.Split(path)
	return id
}
