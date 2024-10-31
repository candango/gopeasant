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
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Transport defines the interface for handling nonce generation and directory
// listing.
type Transport interface {
	// NewNonce generates a new nonce.
	NewNonce() (string, error)
	// NewNonce generates a new nonce.
	Directory() (map[string]interface{}, error)
}

// HttpTransport implements the Transport interface for HTTP communications.
type HttpTransport struct {
	// Url is the base URL for the transport.
	Url string
	// nonceKey is the header key used to retrieve the nonce from responses.
	nonceKey string
	// client is the HTTP client used for making requests.
	client http.Client
}

// NewHttpTransport initializes a new HttpTransport with the given URL and
// nonce key.
func NewHttpTransport(url string, nonceKey string) *HttpTransport {
	return &HttpTransport{
		url,
		nonceKey,
		http.Client{},
	}
}

// Directory returns a map of available resources, including the URL for new
// nonce generation. This method should be overridden if the developer needs
// to retrieve dynamic data from the server's directory.
func (ht *HttpTransport) Directory() (map[string]interface{}, error) {
	return map[string]interface{}{
		"newNonce": ht.Url + "/nonce/new-nonce",
	}, nil
}

// NewNonceUrl returns the URL for generating a new nonce. Developers should
// override this method if the new nonce URL needs to be resolved differently.
func (ht *HttpTransport) NewNonceUrl() (string, error) {
	d, err := ht.Directory()
	if err != nil {
		return "", err
	}
	return d["newNonce"].(string), nil
}

// ResolveNonce extracts the nonce from the response headers using the
// predefined nonceKey. Developers should override this method if the nonce
// needs to be resolved in a different way.
func (ht *HttpTransport) ResolveNonce(res *http.Response) string {
	return res.Header.Get(ht.nonceKey)
}

// NewNonce generates a new nonce by making an HTTP HEAD request to the new
// nonce URL. This method depends on NewNonceUrl and ResolveNonce. The basic
// implementation is provided, but customization should be done in the
// dependent methods. If further customization is needed, developers can use
// this method as a template.
func (ht *HttpTransport) NewNonce() (string, error) {
	url, err := ht.NewNonceUrl()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}
	res, err := ht.client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode > 299 {
		return "", errors.New(res.Status)
	}
	return ht.ResolveNonce(res), nil
}

// Peasant represents an agent in the Peasant protocol, which communicates with
// a bastion.
// It wraps a Transport for handling nonce generation and other communication
// aspects.
type Peasant struct {
	Transport
}

// NewPeasant initializes a new Peasant with the provided Transport.
func NewPeasant(tr Transport) *Peasant {
	return &Peasant{tr}
}

// NewNonce generates a new nonce by delegating the call to the underlying
// Transport.
// This method allows the Peasant to obtain a new nonce for communication with
// a bastion.
func (p *Peasant) NewNonce() (string, error) {
	return p.Transport.NewNonce()
}

// BodyAsString reads the entire body of an HTTP response and returns it as a
// string.
// It consumes the response body, so the caller should not attempt to read from
// it again.
func BodyAsString(res *http.Response) (string, error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// BodyAsJson reads the entire body of an HTTP response and unmarshals it into
// the provided jsonBody.
// It consumes the response body, so the caller should not attempt to read from
// it again.
// The jsonBody parameter should be a pointer to a struct or a slice where JSON
// data will be unmarshaled.
func BodyAsJson(res *http.Response, jsonBody any) error {
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, jsonBody)
	if err != nil {
		return err
	}
	return nil
}
