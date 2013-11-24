// Example usage:
//  ./pivot --target_directory path/to/my/managed/directory some/directory some/other/directory
//
// pivot will search for images in the specified directories and populates the --target_directory.  The directory is searched recursively.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

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

var (
	targetDirectory = flag.String("target_directory", "", "Directory for processed files.")
	test            = flag.Bool("test", false, "Just test")
)

const targetDirectoryEnvVar = "PIVOTDIRECTORY"

func getRepo() string {
	envRepo := os.Getenv(targetDirectoryEnvVar)
	if len(*targetDirectory) != 0 {
		return *targetDirectory
	} else if len(envRepo) != 0 {
		return envRepo
	}
	log.Fatalf("Pivot directory must be set via the %s env var or --target_directory command line argument", targetDirectoryEnvVar)
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
	flag.Parse()
	topLevelFiles := []string{}
	for _, glob := range flag.Args() {
		topLevelFiles = append(topLevelFiles, glob)
	}
	repo := getRepo()
	metadata := internal.FindAllTiffFiles(topLevelFiles)
	fmt.Printf("Found %v files.\n", len(metadata))
	importDirectory := fmt.Sprintf("%s/images/imports", repo)
	for _, m := range metadata {
		exist := doesImageExistInAnyDirectory(m, repo)
		if !exist {
			output := filepath.Join(importDirectory, m.NewFileName())
			fmt.Printf("%s -> %s\n", m.FilePath, output)
			if !*test {
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
