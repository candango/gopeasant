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

package middleware

import (
	"net/http"

	peasant "github.com/candango/gopeasant"
)

// Nonced is a middleware that verifies the presence of a valid nonce in a
// request.
// If the nonce is not provided or is invalid, it prevents the request from
// proceeding.
func Nonced(next http.Handler, s peasant.NonceService) http.Handler {
	return http.HandlerFunc(peasant.NoncedHandlerFunc(s,
		func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		}),
	)
}
