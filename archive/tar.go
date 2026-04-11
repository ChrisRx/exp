package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"runtime"
	"sync/atomic"

	archiveio "go.chrisrx.dev/x/archive/io"
	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/group"
)

var (
	TarMagicHeader = []byte{0x75, 0x73, 0x74, 0x61, 0x72}
	TarMagicOffset = 257
)

func IsTarFile(r io.ReaderAt) bool {
	buf := make([]byte, len(TarMagicHeader))
	n, err := r.ReadAt(buf, int64(TarMagicOffset))
	if err != nil {
		return false
	}
	buf = buf[:n]
	return bytes.Equal(buf[:len(TarMagicHeader)], TarMagicHeader)
}

func UntarFiles(ctx context.Context, r io.Reader, fn func(Reader) error, opts ...Option) error {
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
	defer reader.Close()
	if !IsTarFile(reader) {
		return fmt.Errorf("not a tar file")
	}

	tr := tar.NewReader(reader)

	var n atomic.Uint64
	g := group.New(ctx, group.WithLimit(o.Concurrency))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if ctx.Err() == context.Canceled {
			break
		}

		sr := io.NewSectionReader(reader, reader.Offset(), hdr.Size)
		g.Go(func(ctx context.Context) error {
			if o.Offset != 0 && n.Load() < uint64(o.Offset) {
				n.Add(1)
				return nil
			}

			if o.Limit != 0 && n.Load() >= uint64(o.Limit+o.Offset) {
				cancel()
				return nil
			}
			defer n.Add(1)

			var r = io.Reader(sr)
			if ok, _ := IsGzipFile(sr); ok {
				r, err = gzip.NewReader(sr)
				if err != nil {
					return errors.Stack(err)
				}
			}

			switch hdr.Typeflag {
			case tar.TypeReg:
				return fn(&struct {
					io.Reader
					fs.FileInfo
				}{
					FileInfo: hdr.FileInfo(),
					Reader:   r,
				})
			default:
				return nil
			}
		})
	}
	return g.Wait()
}
