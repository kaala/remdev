package handler

import (
	"net/http"
	"strings"
)

// Auth creates a middleware that checks for a Bearer token.
// If token is empty, all requests pass through.
func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Allow token via header or query param.
			if checkToken(r, token) {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("WWW-Authenticate", `Bearer realm="rdev"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func checkToken(r *http.Request, expected string) bool {
	// Check Authorization header.
	auth := r.Header.Get("Authorization")
	if auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			if strings.TrimPrefix(auth, "Bearer ") == expected {
				return true
			}
		}
	}

	// Check query parameter.
	if r.URL.Query().Get("token") == expected {
		return true
	}

	return false
}
