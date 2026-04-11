package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"

	archiveio "go.chrisrx.dev/x/archive/io"
	"go.chrisrx.dev/x/slices"
)

type Reader interface {
	io.Reader
	fs.FileInfo
}

var GzipMagicHeader = []byte{0x1f, 0x8b}

func IsGzipFile(r io.Reader) (bool, error) {
	buf := make([]byte, len(GzipMagicHeader))
	var n int
	var err error
	switch r := r.(type) {
	case io.ReaderAt:
		n, err = r.ReadAt(buf, 0)
	default:
		n, err = r.Read(buf)
	}
	if err != nil {
		return false, err
	}
	buf = buf[:n]
	return bytes.Equal(buf[:len(GzipMagicHeader)], GzipMagicHeader), nil
}

func ListFiles(r io.Reader, opts ...Option) ([]fs.FileInfo, error) {
	reader, err := archiveio.NewReaderAt(r)
	if err != nil {
		return nil, err
	}
	switch {
	case IsZipFile(reader):
		zr, err := zip.NewReader(reader, reader.Size())
		if err != nil {
			return nil, err
		}
		return slices.Map(zr.File, func(f *zip.File) fs.FileInfo {
			return f.FileInfo()
		}), nil
	case IsTarFile(reader):
		var files []fs.FileInfo
		tr := tar.NewReader(reader)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			files = append(files, hdr.FileInfo())
		}
		return files, nil
	default:
		return nil, fmt.Errorf("not an archive")
	}
}
