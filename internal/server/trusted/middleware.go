// Package trusted contains middleware for checking request by CIDR.
package trusted

import (
	"net"
	"net/http"
)

func Middleware(trustedNetwork string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if trustedNetwork != "" {
			ipStr := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(w, "failed parse ip from http header", http.StatusForbidden)
				return
			}

			_, ipNet, err := net.ParseCIDR(trustedNetwork)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !ipNet.Contains(ip) {
				http.Error(w, "ip not in trusted network", http.StatusForbidden)
				return
			}
		}

		h.ServeHTTP(w, r)
	}
}
