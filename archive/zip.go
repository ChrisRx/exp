package archive

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"iter"
	"runtime"
	"sync/atomic"

	"go.chrisrx.dev/x/errors"
	"go.chrisrx.dev/x/group"
	"go.chrisrx.dev/x/sync"
)

func UnzipJSON[T any](ctx context.Context, r io.Reader, opts ...Option) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		results := sync.NewChan[T](0)
		go func() {
			defer results.Close()

			if err := UnzipFiles(ctx, r, func(data []byte) error {
				var resp T
				if err := json.Unmarshal(data, &resp); err != nil {
					return err
				}
				results.Send(resp)
				return nil
			}, opts...); err != nil {
				var zero T
				yield(zero, err)
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

func UnzipFiles(ctx context.Context, r io.Reader, fn func([]byte) error, opts ...Option) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	o := &options{
		Concurrency: runtime.NumCPU(),
	}
	for _, opt := range opts {
		opt(o)
	}

	reader, size, err := NewReaderAt(r)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(reader, size)
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
		if isGzipCompressed(&zf) {
			zf, err = gzip.NewReader(zf)
			if err != nil {
				return errors.Stack(err)
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
			data, err := io.ReadAll(zf)
			if err != nil {
				return err
			}
			n.Add(1)
			return fn(data)
		})
	}
	return g.Wait()
}
