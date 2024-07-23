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
)

// NonceService defines methods for managing nonces in HTTP requests.
// It provides functionality for blocking, clearing, consuming, getting,
// and checking the provision of nonces.
type NonceService interface {

	// Block blocks the provided HTTP request if the nonce is not valid.
	Block(http.ResponseWriter, *http.Request) error

	// Clear clears a nonce associated with the specified key. If the key
	// doesn't exists no error will be returned.
	//
	// Return errors only if an actual error occours.
	Clear(string) error

	// Consume processes the nonce associated with the specified key and
	// returns a boolean indicating whether the nonce was successfully
	// consumed, along with any error encountered.
	// If nonce connot be consumed header sould be set with the respective http
	// error code.
	Consume(http.ResponseWriter, *http.Request) error

	// GetNonce generates a new nonce, and stores it for a future validation.
	// It returns the nonce as a string and an error if any occurred during
	// the nonce generation or header update.
	GetNonce(*http.Request) (string, error)

	// IsNonced return if the request should be nonced or not.
	IsNonced(*http.Request) bool

	// IsProvided verifies the presence of a valid nonce in the specified HTTP
	// request.
	//
	// If the nonce is not provided or is invalid, it sets the response HTTP
	// status to "Unauthorized", "Forbidden", or another appropriate status
	// based on the specific conditions and checks performed within the method.
	// If nonce is not provided header sould be set with the respective http
	// error code.
	IsProvided(http.ResponseWriter, *http.Request) error
}
