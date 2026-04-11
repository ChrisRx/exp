package archive

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/json/jsontext"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"runtime"
	"sync/atomic"

	archiveio "go.chrisrx.dev/x/archive/io"
	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/group"
	"go.chrisrx.dev/x/sync"
)

var ZipMagicHeader = []byte{0x50, 0x4b, 0x03, 0x04}

func IsZipFile(r io.ReaderAt) bool {
	buf := make([]byte, len(ZipMagicHeader))
	if _, err := r.ReadAt(buf, 0); err != nil {
		return false
	}
	return bytes.Equal(buf, ZipMagicHeader)
}

func UnzipFiles(ctx context.Context, r io.Reader, fn func(Reader) error, opts ...Option) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	o := &options{
		Concurrency: runtime.NumCPU(),
	}
	for _, opt := range opts {
		opt(o)
	}

	reader, err := archiveio.NewReaderAt(r)
	if err != nil {
		return err
	}
	if !IsZipFile(reader) {
		return fmt.Errorf("not a zip file")
	}
	zr, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return err
	}
	if len(zr.File) == 0 {
		return nil
	}

	var n atomic.Uint64
	g := group.New(ctx, group.WithLimit(o.Concurrency))
	for _, file := range zr.File {
		if ctx.Err() == context.Canceled {
			break
		}
		zf, err := file.Open()
		if err != nil {
			return err
		}
		if r, err := file.Open(); err == nil {
			if ok, _ := IsGzipFile(r); ok {
				zf, err = gzip.NewReader(zf)
				if err != nil {
					return errors.Stack(err)
				}
			}
		}
		g.Go(func(ctx context.Context) error {
			defer zf.Close() //nolint:errcheck

			if o.Offset != 0 && n.Load() < uint64(o.Offset) {
				n.Add(1)
				return nil
			}

			if o.Limit != 0 && n.Load() >= uint64(o.Limit+o.Offset) {
				cancel()
				return nil
			}
			defer n.Add(1)
			return fn(&struct {
				io.Reader
				fs.FileInfo
			}{
				FileInfo: file.FileInfo(),
				Reader:   zf,
			})
		})
	}
	return g.MustWait()
}

func UnzipJSON[T any](ctx context.Context, r io.Reader, opts ...Option) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		results := sync.NewChan[T](0)
		go func() {
			defer results.Close()

			if err := UnzipFiles(ctx, r, func(r Reader) error {
				d := jsontext.NewDecoder(r)
				for {
					v, err := d.ReadValue()
					if err != nil {
						if err == io.EOF {
							return nil
						}
						return err
					}
					var resp T
					if err := json.Unmarshal(v, &resp); err != nil {
						return err
					}
					results.Send(resp)
				}
			}, opts...); err != nil {
				yield(*new(T), err)
				return
			}
		}()

		for {
			select {
			case result, ok := <-results.Recv():
				if !ok {
					return
				}
				if !yield(result, nil) {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
}
