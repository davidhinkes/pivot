// Example usage:
//  ./pivot --target_directory path/to/my/managed/directory some/directory some/other/directory
//
// pivot will search for images in the specified directories and populates the --target_directory.  The directory is searched recursively.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/davidhinkes/pivot/internal"
)

func doesFileExist(path string) (bool, error) {
	fn, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	fn.Close()
	return true, nil
}

const targetDirectoryEnvVar = "PIVOTDIRECTORY"

func getRepo(c *cli.Context) string {
	envRepo := os.Getenv(targetDirectoryEnvVar)
	cmdRepo := c.GlobalString("repository")
	if len(cmdRepo) != 0 {
		return cmdRepo
	} else if len(envRepo) != 0 {
		return envRepo
	}
	log.Fatalf("Pivot repo must be set via the %s env var or --repository command line argument", targetDirectoryEnvVar)
	return ""
}

func doesImageExistInAnyDirectory(image internal.Metadata, repo string) bool {
	globExpression := fmt.Sprintf("%s/images/*/%s", repo, image.NewFileName())
	fileNames, err := filepath.Glob(globExpression)
	if err != nil {
		panic(err)
	}
	return len(fileNames) > 0
}

func main() {
	app := cli.NewApp()
	app.Name = "pivot"
	app.Usage = "Manage your photos!"
	app.Flags = []cli.Flag{
		cli.StringFlag{"repository", "", "path to pivot repository"},
		cli.BoolFlag{"test", "just test, don't do anything"},
	}
	app.Commands = []cli.Command{{
		Name:        "import",
		Usage:       "find and import photos recursively from directory",
		Description: "import path/to/some/directory",
		Action:      importCommand}}
	app.Run(os.Args)
}

func importCommand(context *cli.Context) {
	topLevelFiles := []string{}
	for _, glob := range context.Args() {
		topLevelFiles = append(topLevelFiles, glob)
	}
	repo := getRepo(context)
	metadata := internal.FindAllTiffFiles(topLevelFiles)
	fmt.Printf("Found %v files.\n", len(metadata))
	importDirectory := fmt.Sprintf("%s/images/imports", repo)
	for _, m := range metadata {
		exist := doesImageExistInAnyDirectory(m, repo)
		if !exist {
			output := filepath.Join(importDirectory, m.NewFileName())
			fmt.Printf("%s -> %s\n", m.FilePath, output)
			if !context.GlobalBool("test") {
				os.MkdirAll(filepath.Dir(output), 0700)
				outputFile, _ := os.Create(output)
				f, _ := os.Open(m.FilePath)
				io.Copy(outputFile, f)
			}
		} else {
			fmt.Printf("%s already present\n", m.NewFileName())
		}
	}
}
