package gzip

import (
	"compress/gzip"
	"io"
)

// Reader describes wrapper for standard reader with gzip uncompress support.
type Reader struct {
	r  io.ReadCloser
	zr gzip.Reader
}

// NewReader creates new instance of wrapped reader.
func NewReader(r io.ReadCloser) (*Reader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &Reader{r: r, zr: *zr}, nil
}

// Read reads up to len(p) uncompressed bytes.
func (gr *Reader) Read(p []byte) (n int, err error) {
	return gr.zr.Read(p)
}

// Close closes wrapped reader.
func (gr *Reader) Close() error {
	if err := gr.zr.Close(); err != nil {
		return err
	}

	return gr.r.Close()
}
