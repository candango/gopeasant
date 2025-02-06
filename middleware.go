package peasant

import (
	"context"
	"net/http"
)

// Nonced is a middleware that verifies the presence of a valid nonce in a
// request.
// If the nonce is not provided or is invalid, it prevents the request from
// proceeding.
func Nonced(next http.Handler, s NonceService) http.Handler {
	return http.HandlerFunc(NoncedHandlerFunc(s,
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}),
	)
}

// NonceServed creates a middleware that injects a NonceService into the
// request's context.
func NonceServed(s NonceService, key string) func(http.Handler) http.Handler {
	if key == "" {
		key = "nonce-service"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), key, s)
			req := r.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}
