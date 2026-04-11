package io

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func BinarySearchSize(r io.ReaderAt) (int64, error) {
	buf := make([]byte, 1)
	lo, hi := int64(0), int64(1)
	for {
		_, err := r.ReadAt(buf, hi-1)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		lo = hi
		hi *= 2
	}

	for lo < hi {
		n := int64(uint(lo+hi) >> 1)
		switch _, err := r.ReadAt(buf, n); err {
		case io.EOF:
			hi = n
		case nil:
			lo = n + 1
		default:
			return 0, err
		}
	}
	return lo, nil
}

func Size(r io.Reader) (_ int64, reterr error) {
	switch r := r.(type) {
	case *bytes.Buffer:
		return int64(r.Len()), nil
	case *bytes.Reader:
		return int64(r.Len()), nil
	case *strings.Reader:
		return int64(r.Len()), nil
	case *io.SectionReader:
		return r.Size(), nil
	case *os.File:
		fi, err := r.Stat()
		if err != nil {
			return 0, err
		}
		return fi.Size(), nil
	case io.Seeker:
		cur, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return 0, err
		}
		end, err := r.Seek(0, io.SeekEnd)
		if err != nil {
			return 0, err
		}
		start, err := r.Seek(0, io.SeekStart)
		if err != nil {
			return 0, err
		}
		pos, err := r.Seek(cur, io.SeekStart)
		if err != nil {
			return 0, fmt.Errorf("cannot restore position: %w", err)
		}
		if pos != cur {
			return 0, fmt.Errorf("restored position mismatch: %d != %d", cur, pos)
		}
		return end - start, nil
	case io.ReaderAt:
		return BinarySearchSize(r)
	}
	return 0, fmt.Errorf("cannot determine size for %T", r)
}
