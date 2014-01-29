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
		Action:      importCommand,
		Flags: []cli.Flag{
			cli.BoolFlag{"remove", "remove source files onced copied"},
		}}}
	app.Run(os.Args)
}

func copyFile(inputPath, outputPath string) {
	os.MkdirAll(filepath.Dir(outputPath), 0700)
	outputFile, err := os.Create(outputPath)
	defer outputFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	inputFile, err := os.Open(inputPath)
	defer inputFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		log.Fatal(err)
	}
}

func importCommand(context *cli.Context) {
	if context.GlobalBool("test") {
		fmt.Println("Just a test, not importing anything.")
	}
	topLevelFiles := []string{}
	for _, glob := range context.Args() {
		topLevelFiles = append(topLevelFiles, glob)
	}
	repo := getRepo(context)
	metadata := internal.FindAllTiffFiles(topLevelFiles)
	fmt.Printf("Found %v file(s).\n", len(metadata))
	importDirectory := fmt.Sprintf("%s/images/vault", repo)
	imagesImported := 0
	for _, m := range metadata {
		exist := doesImageExistInAnyDirectory(m, repo)
		if !exist {
			imagesImported = imagesImported + 1
			if !context.GlobalBool("test") {
				output := filepath.Join(importDirectory, m.NewFileName())
				copyFile(m.FilePath, output)
				if context.Bool("remove") {
					os.Remove(m.FilePath)
				}
			}
		}
	}
	if context.GlobalBool("test") {
		fmt.Printf("Would have imported %v file(s), had this not been a test.\n", imagesImported)
	} else {
		fmt.Printf("Imported %v file(s).\n", imagesImported)
	}
}
