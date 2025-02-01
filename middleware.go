package peasant

import (
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
