package peasant

import (
	"net/http"

	"github.com/candango/httpok"
)

func NoncedHandlerFunc(
	s NonceService, f func(http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Skip(r) {
			f(w, r)
			return
		}
		wrapped := &httpok.WrappedWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		err := s.Provided(wrapped, r)
		if err != nil {
			wrapped.WriteHeader(http.StatusInternalServerError)
			return
		}
		if wrapped.StatusCode >= 300 {
			return
		}
		err = s.Consume(wrapped, r)
		if err != nil {
			wrapped.WriteHeader(http.StatusInternalServerError)
			return
		}
		if wrapped.StatusCode >= 300 {
			return
		}
		nonce, err := s.GetNonce(r)
		if err != nil {
			wrapped.WriteHeader(http.StatusInternalServerError)
			return
		}
		if wrapped.StatusCode >= 300 {
			return
		}
		wrapped.Header().Add("nonce", nonce)
		f(wrapped, r)
	}
}
