package archive

import (
	"bytes"
	"io"
	"log/slog"

	"go.chrisrx.dev/x/sync"
)

type readerAt struct {
	mu sync.Mutex
	r  io.ReadSeeker
}

func NewReaderAt(r io.Reader) (io.ReaderAt, int64, error) {
	switch r := r.(type) {
	case io.ReadSeeker:
		size, err := r.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, 0, err
		}
		return &readerAt{r: r}, size, nil
	case io.ReaderAt:
		var size int64
		for {
			buf := make([]byte, 1024)
			n, err := r.ReadAt(buf, 0)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, 0, err
			}
			size += int64(n)
		}
		return r, size, nil
	default:
		slog.Warn("reader does not implement io.ReaderAt")
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, 0, err
		}
		return bytes.NewReader(data), int64(len(data)), nil
	}
}

func (r *readerAt) ReadAt(p []byte, offset int64) (n int, reterr error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	orig, err := r.r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	defer func() {
		if reterr == nil {
			if _, err := r.r.Seek(orig, io.SeekStart); err != nil {
				reterr = err
			}
		}
	}()

	if _, err := r.r.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

type readCloser struct {
	io.Reader
	closeFunc func() error
}

func (r *readCloser) Close() error { return r.closeFunc() }

func peek(r *io.ReadCloser, n int) []byte {
	buf := make([]byte, n)
	if _, err := io.ReadFull(*r, buf); err != nil {
		return nil
	}
	*r = &readCloser{
		Reader:    io.MultiReader(bytes.NewReader(buf), *r),
		closeFunc: (*r).Close,
	}
	return buf
}
