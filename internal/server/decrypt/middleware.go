// Package decrypt describes middleware to decrypt request body with RSA private key
package decrypt

import (
	"crypto/rsa"
	"net/http"
)

func Middleware(privateKey *rsa.PrivateKey, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if privateKey != nil {
			aesReader, err := NewAESReader(privateKey, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = aesReader
		}

		h.ServeHTTP(w, r)
	}
}
