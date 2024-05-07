package gzip

import (
	"compress/gzip"
	"io"
)

type Reader struct {
	r  io.ReadCloser
	zr gzip.Reader
}

func NewReader(r io.ReadCloser) (*Reader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &Reader{r: r, zr: *zr}, nil
}

func (gr *Reader) Read(p []byte) (n int, err error) {
	return gr.zr.Read(p)
}

func (gr *Reader) Close() error {
	if err := gr.zr.Close(); err != nil {
		return err
	}

	return gr.r.Close()
}
