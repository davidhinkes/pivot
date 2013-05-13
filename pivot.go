// Example usage:
//  ./pivot --target_directory path/to/my/managed/directory glob/style/path
package main

import (
	"crypto"
	_ "crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/rwcarlsen/goexif/tiff"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type imageMetadata struct {
	originalFileName string
	date             string
	hash             string
}

func (i imageMetadata) NewFileName() string {
	return i.hash + strings.ToLower(filepath.Ext(i.originalFileName))
}

func main() {
	var targetDirectory = flag.String("target_directory", "", "Directory for processed files.")
	var test = flag.Bool("test", false, "Just test")
	flag.Parse()
	for _, glob := range flag.Args() {
		files, _ := filepath.Glob(glob)
		for _, fn := range files {
			m, err := extractTiffMetadata(fn)
			if err == nil {
				path := filepath.Join(*targetDirectory, m.date)
				output := filepath.Join(path, m.NewFileName())
				fmt.Printf("%s -> %s\n", m.originalFileName, output)
				if !*test {
					os.MkdirAll(path, 0700)
					outputFile, _ := os.Create(output)
					f, _ := os.Open(fn)
					io.Copy(outputFile, f)
				}
			}
		}
	}
}

const rawTimeFormat = "2006:01:02 15:04:05"
const pivotDateFormat = "20060102"

func extractTiffMetadata(fn string) (imageMetadata, error) {
	f, err := os.Open(fn)
	if err != nil {
		return imageMetadata{}, err
	}
	x, err := tiff.Decode(f)
	if err != nil {
		return imageMetadata{}, err
	}
	for _, dir := range x.Dirs {
		for _, tag := range dir.Tags {
			if tag.Id == 306 {
				t, err := time.Parse(rawTimeFormat, tag.StringVal())
				if err != nil {
					return imageMetadata{}, err
					log.Fatal(err)
				}
				dir := t.Format(pivotDateFormat)
				h := crypto.SHA256.New()
				io.WriteString(h, fn)
				newFileName := hex.EncodeToString(h.Sum(nil))
				return imageMetadata{fn, dir, newFileName}, nil
			}
		}
	}
	return imageMetadata{}, errors.New("No time data found in tiff")
}
