// Copyright 2023-2024 Flavio Garcia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
