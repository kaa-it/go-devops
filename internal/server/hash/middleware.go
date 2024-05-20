package hash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
)

func Middleware(key string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := r.Header.Get("Hash")

		if key == "" || hash == "" {
			h.ServeHTTP(w, r)
			return
		}

		// TODO: Handle calculate hash at write

		body, err := io.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		decodedHash, err := base64.StdEncoding.DecodeString(hash)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hm := hmac.New(sha256.New, []byte(key))
		hm.Write(body)

		calculatedHash := hm.Sum(nil)

		if !bytes.Equal(decodedHash, calculatedHash) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		h.ServeHTTP(w, r)
	}
}
