package hash

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
)

func Middleware(key string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := r.Header.Get("Hash")

		if key == "" || hash == "" || strings.ToLower(hash) == "none" {
			h.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()

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
