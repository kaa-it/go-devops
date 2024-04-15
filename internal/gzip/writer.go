package gzip

import (
	"compress/gzip"
	"net/http"
)

var contentTypesForGzip = []string{
	"text/html",
	"application/json",
}

type Writer struct {
	w        http.ResponseWriter
	zw       *gzip.Writer
	compress bool
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{w: w, zw: gzip.NewWriter(w), compress: false}
}

func (gw *Writer) Header() http.Header {
	return gw.w.Header()
}

func (gw *Writer) WriteHeader(statusCode int) {
	contentType := gw.w.Header().Get("Content-Type")

	if statusCode < 300 && isValidContentType(contentType) {
		gw.w.Header().Set("Content-Encoding", "gzip")
		gw.compress = true
	}

	gw.w.WriteHeader(statusCode)
}

func (gw *Writer) Write(b []byte) (int, error) {
	if gw.compress {
		return gw.zw.Write(b)
	}

	return gw.w.Write(b)
}

func (gw *Writer) Close() error {
	return gw.zw.Close()
}

func isValidContentType(contentType string) bool {
	for _, ct := range contentTypesForGzip {
		if ct == contentType {
			return true
		}
	}

	return false
}
