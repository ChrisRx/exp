package archive_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"sync"
	"testing"

	"go.chrisrx.dev/x/archive"
	"go.chrisrx.dev/x/assert"
)

func TestUntarFiles(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		ctx := t.Context()

		f, err := os.Open("testdata/simple.json.tar")
		if err != nil {
			t.Fatal(err)
		}
		var mu sync.Mutex
		var files []File
		if err := archive.UntarFiles(ctx, f, func(r archive.Reader) error {
			data, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			var file File
			if err := json.Unmarshal(data, &file); err != nil {
				return err
			}
			mu.Lock()
			files = append(files, file)
			mu.Unlock()
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		assert.ElementsMatch(t, []File{
			{
				ID:     1,
				Name:   "testing",
				Values: []int{1, 2, 3, 4, 5},
			},
			{
				ID:     2,
				Name:   "testing2",
				Values: []int{5, 6, 7, 8, 9},
			},
			{
				ID:     3,
				Name:   "testing3",
				Values: []int{9, 10, 11, 12, 13},
			},
		}, files)
	})

	t.Run("simple with gzip", func(t *testing.T) {
		ctx := t.Context()

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
		var mu sync.Mutex
		var files []File
		if err := archive.UntarFiles(ctx, bytes.NewReader(data), func(r archive.Reader) error {
			data, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			var file File
			if err := json.Unmarshal(data, &file); err != nil {
				return err
			}
			mu.Lock()
			files = append(files, file)
			mu.Unlock()
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		assert.ElementsMatch(t, []File{
			{
				ID:     1,
				Name:   "testing",
				Values: []int{1, 2, 3, 4, 5},
			},
			{
				ID:     2,
				Name:   "testing2",
				Values: []int{5, 6, 7, 8, 9},
			},
			{
				ID:     3,
				Name:   "testing3",
				Values: []int{9, 10, 11, 12, 13},
			},
		}, files)
	})
}

type File struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Values []int  `json:"values"`
}
