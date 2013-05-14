// Example usage:
//  ./pivot --target_directory path/to/my/managed/directory glob/style/path...
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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type imageMetadata struct {
	date string
	hash string
}

func (i imageMetadata) NewFileName(oldFileName string) string {
	return i.hash + strings.ToLower(filepath.Ext(oldFileName))
}

func newImageMetadata(filePath string, meta *imageMetadata) bool {
	f, err := os.Open(filePath)
	defer f.Close()
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		panic(err)
	}
	*meta, err = extractTiffMetadata(f)
	if err != nil {
		return false
	}
	return true
}

func main() {
	var targetDirectory = flag.String("target_directory", "", "Directory for processed files.")
	var test = flag.Bool("test", false, "Just test")
	flag.Parse()
	for _, glob := range flag.Args() {
		files, _ := filepath.Glob(glob)
		for _, fn := range files {
			var m imageMetadata
			if !newImageMetadata(fn, &m) {
				continue
			}
			path := filepath.Join(*targetDirectory, m.date)
			output := filepath.Join(path, m.NewFileName(fn))
			exist, err := doesFileExist(output)
			if err != nil {
				panic(err)
			}
			if !exist {
				fmt.Printf("%s -> %s\n", fn, output)
				if !*test {
					os.MkdirAll(path, 0700)
					outputFile, _ := os.Create(output)
					f, _ := os.Open(fn)
					io.Copy(outputFile, f)
				}
			} else {
				fmt.Printf("%s already present\n", output)
			}
		}
	}
}

func doesFileExist(path string) (bool, error) {
	fn, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	fn.Close()
	return true, err
}

const rawTimeFormat = "2006:01:02 15:04:05"
const pivotDateFormat = "20060102"

func extractTiffMetadata(f *os.File) (imageMetadata, error) {
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
				}
				date := t.Format(pivotDateFormat)
				h := crypto.SHA256.New()
				f.Seek(0, 0)
				contents, err := ioutil.ReadAll(f)
				if err != nil {
					panic(err)
				}
				io.WriteString(h, string(contents))
				newFileName := hex.EncodeToString(h.Sum(nil))
				return imageMetadata{date, newFileName}, nil
			}
		}
	}
	return imageMetadata{}, errors.New("No time data found in tiff")
}
