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
		cli.BoolFlag{
			Name:  "pre-dump",
			Usage: "Perform a pre-dump to minimize downtime",
		},
	},
	Action: func(c *cli.Context) {
		srcUrl := validate.ParseURL(c.String("src"))
		dstUrl := validate.ParseURL(c.String("dst"))

		log.Println("Performing validations")
		src, dst := validate.Validate(srcUrl, dstUrl)

		log.Println("Preparing everything to do a checkpoint")
		containerId := getContainerId(srcUrl.Path)
		var imagesPath string

		if c.Bool("pre-dump") {
			// Process pre-dump
			predumpPath := fmt.Sprintf("%s/images/0", srcUrl.Path)
			prepareDir(src, predumpPath)

			checkpoint(src, containerId, predumpPath, true)

			srcTarFile := fmt.Sprintf("%s/predump.tar.gz", srcUrl.Path)
			prepareTar(src, srcTarFile, predumpPath)

			prepareDir(dst, fmt.Sprintf("%s/images/0", dstUrl.Path))

			log.Println("Copying predump image to dst")
			err := cmd.Scp(src.URL(srcTarFile), dst.URL(fmt.Sprintf("%s/images/0", dstUrl.Path)))
			if err != nil {
				log.Fatal("Error copying predump image files to dst", err)
			}

			dstTarFile := fmt.Sprintf("%s/images/0/predump.tar.gz", dstUrl.Path)
			unpackTar(dst, dstTarFile, fmt.Sprintf("%s/images/0", dstUrl.Path))

			// Process final image
			imagesPath = fmt.Sprintf("%s/images/1", srcUrl.Path)
			log.Println("Performing the checkpoint")
			_, _, err = src.Run("sudo", "runc", "--id", containerId, "checkpoint", "--image-path", imagesPath, "--prev-images-dir", predumpPath)
			if err != nil {
				log.Fatal("Error performing checkpoint:", err)
			}

			srcTarFile = fmt.Sprintf("%s/dump.tar.gz", srcUrl.Path)
			prepareTar(src, srcTarFile, imagesPath)
			prepareDir(dst, fmt.Sprintf("%s/images/1", dstUrl.Path))

			log.Println("Copying predump image to dst")
			err = cmd.Scp(src.URL(srcTarFile), dst.URL(fmt.Sprintf("%s/images/1", dstUrl.Path)))
			if err != nil {
				log.Fatal("Error copying predump image files to dst", err)
			}

			dstTarFile = fmt.Sprintf("%s/images/1/dump.tar.gz", dstUrl.Path)
			unpackTar(dst, dstTarFile, fmt.Sprintf("%s/images/1", dstUrl.Path))

			log.Println("Performing the restore")
			configFilePath := fmt.Sprintf("%s/config.json", dstUrl.Path)
			runtimeFilePath := fmt.Sprintf("%s/runtime.json", dstUrl.Path)
			dstImagesPath := fmt.Sprintf("%s/images/1", dstUrl.Path)
			err = dst.Start("sudo", "runc", "--id", containerId, "restore", "--image-path", dstImagesPath, "--config-file", configFilePath, "--runtime-file", runtimeFilePath)
			if err != nil {
				log.Fatal("Error performing restore:", err)
			}

		} else {
			imagesPath = fmt.Sprintf("%s/images", srcUrl.Path)
			prepareDir(src, imagesPath)

			checkpoint(src, containerId, imagesPath, false)

			srcTarFile := fmt.Sprintf("%s/dump.tar.gz", srcUrl.Path)
			prepareTar(src, srcTarFile, imagesPath)

			prepareDir(dst, fmt.Sprintf("%s/images", dstUrl.Path))

			log.Println("Copying checkpoint image to dst")
			err := cmd.Scp(src.URL(srcTarFile), dst.URL(fmt.Sprintf("%s/images", dstUrl.Path)))
			if err != nil {
				log.Fatal("Error copying image files to dst", err)
			}

			dstTarFile := fmt.Sprintf("%s/images/dump.tar.gz", dstUrl.Path)
			unpackTar(dst, dstTarFile, fmt.Sprintf("%s/images", dstUrl.Path))

			log.Println("Performing the restore")
			configFilePath := fmt.Sprintf("%s/config.json", dstUrl.Path)
			runtimeFilePath := fmt.Sprintf("%s/runtime.json", dstUrl.Path)
			dstImagesPath := fmt.Sprintf("%s/images", dstUrl.Path)
			err = dst.Start("sudo", "runc", "--id", containerId, "restore", "--image-path", dstImagesPath, "--config-file", configFilePath, "--runtime-file", runtimeFilePath)
			if err != nil {
				log.Fatal("Error performing restore:", err)
			}

		}

		log.Println("Restore completed")

	},
}

func unpackTar(cmd cmd.Cmd, tarFile, workDir string) {
	log.Println("Preparing image at destination host")
	_, _, err := cmd.Run("sudo", "tar", "-C", workDir, "-xvzf", tarFile)
	if err != nil {
		log.Fatal("Error uncompressing image in destination:", err)
	}
}

func prepareTar(cmd cmd.Cmd, tarFile, workDir string) {
	_, _, err := cmd.Run("sudo", "tar", "-czf", tarFile, "-C", fmt.Sprintf("%s/", workDir), ".")
	if err != nil {
		log.Fatal("Error compressing image in source:", err)
	}
}

func checkpoint(cmd cmd.Cmd, containerId, imagesPath string, predump bool) {
	log.Printf("Performing the checkpoint predump = %t\n", predump)
	args := []string{"runc", "--id", containerId, "checkpoint", "--image-path", imagesPath}
	if predump {
		args = append(args, "--pre-dump")
	}
	_, _, err := cmd.Run("sudo", args...)
	if err != nil {
		log.Fatal("Error performing checkpoint:", err)
	}
}

func prepareDir(cmd cmd.Cmd, path string) {
	_, _, err := cmd.Run("mkdir", "-p", path)
	if err != nil {
		log.Fatal("Error preparing pre-dump dir:", err)
	}
}

func getContainerId(path string) string {
	_, id := filepath.Split(path)
	return id
}
