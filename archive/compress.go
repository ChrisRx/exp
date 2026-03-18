package archive

import (
	"bytes"
	"io"
)

var gzipMagicHeader = []byte{'\x1f', '\x8b'}

func isGzipCompressed(r *io.ReadCloser) bool {
	return bytes.Equal(
		peek(r, len(gzipMagicHeader)),
		gzipMagicHeader,
	)
}
