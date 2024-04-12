package gzip

import (
	"net/http"
	"strings"
)

func Middleware(h http.HandlerFunc) http.HandlerFunc {
	var gw *Writer
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		supportGzip := isSupportGzip(r.Header.Values("Accept-Encoding"))
		if supportGzip {
			if gw == nil {
				gw = NewWriter(w)
			}

			ow = gw

			defer func() {
				gw.Close()
				gw.Reset()
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
