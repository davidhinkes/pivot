package internal

import (
	"crypto"
	_ "crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/tiff"
)

type queue struct {
	i     int
	items []interface{}
}

func (q *queue) Push(i interface{}) {
	q.items = append(q.items, i)
}

func (q *queue) Pop() interface{} {
	if q.IsEmpty() {
		panic("Pop called on empty queue")
	}
	v := q.items[q.i]
	q.i++
	return v
}

func (q queue) IsEmpty() bool {
	return !(q.i < len(q.items))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// Metadata contains image metadata.
type Metadata struct {
	FilePath string
	Date     string
	Hash     string
}

// Computes new pivot file name for image.
func (i Metadata) NewFileName(oldFileName string) string {
	return i.Hash + strings.ToLower(filepath.Ext(oldFileName))
}

const rawTimeFormat = "2006:01:02 15:04:05"
const pivotDateFormat = "20060102"

func extractTiffMetadata(filePath string) (Metadata, error) {
	f, err := os.Open(filePath)
	panicOnError(err)
	defer f.Close()
	x, err := tiff.Decode(f)
	if err != nil {
		return Metadata{}, err
	}
	for _, dir := range x.Dirs {
		for _, tag := range dir.Tags {
			if tag.Id == 306 {
				t, err := time.Parse(rawTimeFormat, tag.StringVal())
				if err != nil {
					return Metadata{}, err
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
				return Metadata{filePath, date, newFileName}, nil
			}
		}
	}
	return Metadata{}, errors.New("No time data found in tiff")
}

func FindAllTiffFiles(topLevelDirs []string) []Metadata {
	result := []Metadata{}
	var explore queue
	for _, dir := range topLevelDirs {
		explore.Push(dir)
	}
	for !explore.IsEmpty() {
		fileName := explore.Pop().(string)
		file, err := os.Open(fileName)
		panicOnError(err)
		info, err := file.Stat()
		panicOnError(err)
		if info.IsDir() {
			// This is a directory.  Add children.
			names, err := file.Readdirnames(0)
			panicOnError(err)
			for _, name := range names {
				explore.Push(filepath.Join(fileName, name))
			}
		} else {
			// This is a real file; try to process it.
			md, err := extractTiffMetadata(fileName)
			if err == nil {
				result = append(result, md)
			}
		}
		file.Close()
	}
	return result
}
