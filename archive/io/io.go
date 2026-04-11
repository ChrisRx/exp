package io

import (
	"fmt"
	"io"
	"sync"
)

type readerFromReaderAt struct {
	r io.ReaderAt
}

func ReaderFromReaderAt(r io.ReaderAt) io.Reader {
	return &readerFromReaderAt{r}
}

func (r *readerFromReaderAt) Read(p []byte) (int, error) {
	return r.r.ReadAt(p, 0)
}

type readerAtFromReadSeeker struct {
	mu sync.Mutex
	rs io.ReadSeeker
}

func ReaderAtFromReadSeeker(rs io.ReadSeeker) io.ReaderAt {
	return &readerAtFromReadSeeker{rs: rs}
}

func (r *readerAtFromReadSeeker) ReadAt(p []byte, offset int64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	orig, err := r.rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	if _, err := r.rs.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}
	n, err := r.rs.Read(p)
	if err != nil {
		return n, err
	}
	if _, err := r.rs.Seek(orig, io.SeekStart); err != nil {
		return n, err
	}
	return n, nil
}

type ReaderAt struct {
	mu           sync.Mutex
	r            io.ReaderAt
	offset, size int64

	closeFn func() error
}

func NewReaderAt(r io.Reader) (*ReaderAt, error) {
	ar := &ReaderAt{
		closeFn: func() error {
			return nil
		},
	}
	if err := ar.Reset(r); err != nil {
		return nil, err
	}
	return ar, nil
}

func (ar *ReaderAt) Reset(r io.Reader) error {
	switch r := r.(type) {
	case io.ReadSeeker:
		size, err := Size(r)
		if err != nil {
			return err
		}
		ar.r = ReaderAtFromReadSeeker(r)
		cur, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		ar.offset = cur
		ar.size = size
	case io.ReaderAt:
		size, err := BinarySearchSize(r)
		if err != nil {
			return err
		}
		ar.r = r
		ar.offset = 0
		ar.size = size
	default:
		return fmt.Errorf("reader does not implement io.ReaderAt")
	}

	if rc, ok := r.(io.Closer); ok {
		ar.closeFn = rc.Close
	} else {
		ar.closeFn = func() error {
			return nil
		}
	}
	return nil
}

func (r *ReaderAt) Close() error {
	return r.closeFn()
}

func (r *ReaderAt) Offset() int64 {
	return r.offset
}

func (r *ReaderAt) Size() int64 {
	return r.size
}

func (r *ReaderAt) Peek(n int) ([]byte, error) {
	buf := make([]byte, n)
	n, err := r.r.ReadAt(buf, 0)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (r *ReaderAt) Read(p []byte) (int, error) {
	n, err := r.r.ReadAt(p, r.offset)
	r.offset += int64(n)
	return n, err
}

func (r *ReaderAt) ReadAt(p []byte, offset int64) (n int, reterr error) {
	return r.r.ReadAt(p, offset)
}
