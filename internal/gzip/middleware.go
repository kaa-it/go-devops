// Package gzip provides support for compressing agent reports and server responses in gzip format.
package gzip

import (
	"net/http"
	"strings"
)

// Middleware wraps request handler to uncompress agent reports at server and compress server responses.
//
// Compresses server response if agent sends gzip in Accept-Encoding request header.
// Uncompresses client report if agent request contains gzip in Content-Encoding request header.
func Middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		supportGzip := isSupportGzip(r.Header.Values("Accept-Encoding"))
		if supportGzip {
			var gw = NewWriter(w)

			ow = gw

			defer func() {
				gw.Close()
			}()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			gr, err := NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = gr
			defer gr.Close()
		}

		h.ServeHTTP(ow, r)
	}
}

func isSupportGzip(acceptEncodings []string) bool {
	for _, encoding := range acceptEncodings {
		if strings.Contains(encoding, "gzip") {
			return true
		}
	}

	return false
}
