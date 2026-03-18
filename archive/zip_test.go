package archive_test

import (
	"bytes"
	"os"
	"testing"

	"go.chrisrx.dev/x/archive"
	"go.chrisrx.dev/x/assert"
)

func TestUnzip(t *testing.T) {
	ctx := t.Context()

	data, err := os.ReadFile("testdata/simple.json.gz.zip")
	if err != nil {
		t.Fatal(err)
	}
	type File struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Values []int  `json:"values"`
	}
	var files []File
	for v, err := range archive.UnzipJSON[File](ctx, bytes.NewReader(data)) {
		if err != nil {
			t.Error(err)
		}
		files = append(files, v)
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
}
