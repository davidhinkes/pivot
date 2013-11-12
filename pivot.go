// Example usage:
//  ./pivot --target_directory path/to/my/managed/directory some/directory some/other/directory
//
// pivot will search for images in the specified directories and populates the --target_directory.  The directory is searched recursively. 
package main

import (
	"flag"
	"fmt"
	"io"
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

func main() {
	var targetDirectory = flag.String("target_directory", "", "Directory for processed files.")
	var test = flag.Bool("test", false, "Just test")
	flag.Parse()
	topLevelFiles := []string{}
	for _, glob := range flag.Args() {
		topLevelFiles = append(topLevelFiles, glob)
	}
	metadata := internal.FindAllTiffFiles(topLevelFiles)
	fmt.Printf("Found %v files.\n", len(metadata))
	for _, m := range metadata {
		path := filepath.Join(*targetDirectory, m.Date)
		output := filepath.Join(path, m.NewFileName(m.FilePath))
		exist, err := doesFileExist(output)
		if err != nil {
			panic(err)
		}
		if !exist {
			fmt.Printf("%s -> %s\n", m.FilePath, output)
			if !*test {
				os.MkdirAll(path, 0700)
				outputFile, _ := os.Create(output)
				f, _ := os.Open(m.FilePath)
				io.Copy(outputFile, f)
			}
		} else {
			fmt.Printf("%s already present\n", output)
		}
	}
}
