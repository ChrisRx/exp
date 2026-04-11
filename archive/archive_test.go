package archive

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"go.chrisrx.dev/x/assert"
	"go.chrisrx.dev/x/slices"
)

func TestList(t *testing.T) {
	t.Run("tar", func(t *testing.T) {
		f, err := os.Open("testdata/simple.json.tar")
		if err != nil {
			t.Fatal(err)
		}
		files, err := ListFiles(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			fmt.Printf("%v\n", file)
		}
		assert.ElementsMatch(t, []string{
			"file1.json.gz",
			"file2.json.gz",
			"file3.json.gz",
		}, slices.Map(files, func(fi fs.FileInfo) string {
			return fi.Name()
		}))
	})

	t.Run("tar with gzip", func(t *testing.T) {
		f, err := os.Open("testdata/simple.json.tar.gz")
		if err != nil {
			t.Fatal(err)
		}
		gr, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(gr)
		if err != nil {
			t.Fatal(err)
		}
		files, err := ListFiles(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			fmt.Printf("%v\n", file)
		}
		assert.ElementsMatch(t, []string{
			"file1.json.gz",
			"file2.json.gz",
			"file3.json.gz",
		}, slices.Map(files, func(fi fs.FileInfo) string {
			return fi.Name()
		}))
	})

	t.Run("zip", func(t *testing.T) {
		f, err := os.Open("testdata/simple.json.gz.zip")
		if err != nil {
			t.Fatal(err)
		}
		files, err := ListFiles(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			fmt.Printf("%v\n", file)
		}
		assert.ElementsMatch(t, []string{
			"file1.json.gz",
			"file2.json.gz",
			"file3.json.gz",
		}, slices.Map(files, func(fi fs.FileInfo) string {
			return fi.Name()
		}))
	})
}
