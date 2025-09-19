package peasant

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Transport defines the interface for handling nonce generation and directory
// listing.
type Transport interface {
	// NewNonce generates a new nonce.
	NewNonce() (string, error)
	// Directory returns the directory map.
	Directory() (map[string]any, error)
}

// HttpTransport implements the Transport interface for HTTP communications.
type HttpTransport struct {
	// DirectoryProvider is embedded to provide directory-related functionality.
	DirectoryProvider
	// DirectoryKey is the key used to retrieve the directory-related url.
	DirectoryKey string
	// DirectoryMethod specifies the HTTP method used for directory operations.
	DirectoryMethod string
	// Client is the HTTP Client used for making requests.
	http.Client
	// NonceKey is the header key used to retrieve the nonce from responses.
	NonceKey string
}

// NewHttpTransport initializes and returns a new HttpTransport using the
// provided DirectoryProvider. It sets up the transport in the provider and
// returns the configured HttpTransport. Returns an error if setting the
// transport fails.
func NewHttpTransport(p DirectoryProvider) (*HttpTransport, error) {
	ht := &HttpTransport{
		DirectoryProvider: p,
		DirectoryKey:      "newNonce",
		DirectoryMethod:   http.MethodHead,
		Client:            http.Client{},
		NonceKey:          "Nonce",
	}
	if err := p.SetTransport(ht); err != nil {
		return nil, err
	}
	return ht, nil
}

// NewNonceUrl returns the URL for generating a new nonce. Developers should
// override this method if the new nonce URL needs to be resolved differently.
func (ht *HttpTransport) NewNonceUrl() (string, error) {
	d, err := ht.Directory()
	if err != nil {
		return "", err
	}
	val, ok := d[ht.DirectoryKey]
	if !ok {
		return "", fmt.Errorf(
			"the transport wasn't able to find the nonce key %s",
			ht.DirectoryKey)
	}
	return val.(string), nil
}

// ResolveNonce extracts the nonce from the response headers using the
// predefined nonceKey. Developers should override this method if the nonce
// needs to be resolved in a different way.
func (ht *HttpTransport) ResolveNonce(res *http.Response) string {
	return res.Header.Get(ht.NonceKey)
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
	req, err := http.NewRequest(ht.DirectoryMethod, url, nil)
	if err != nil {
		return "", err
	}
	res, err := ht.Client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode > 299 {
		return "", errors.New(res.Status)
	}
	return ht.ResolveNonce(res), nil
}

// DirectoryProvider defines the interface for objects that provide directory
// information. Implementations should support retrieving a directory map,
// getting the URL, and setting the transport.
type DirectoryProvider interface {
	// Directory fetches the directory map from the provider.
	Directory() (map[string]any, error)
	// GetUrl returns the provider's URL.
	GetUrl() string
	// SetTransport sets the transport mechanism for the provider.
	SetTransport(Transport) error
}

// MemoryDirectoryProvider is an in-memory implementation of DirectoryProvider.
// It holds a URL and returns predefined directory data.
type MemoryDirectoryProvider struct {
	url string
}

// Directory returns a static map containing the directory endpoints.
// It constructs the "newNonce" endpoint using the provider's URL.
func (p *MemoryDirectoryProvider) Directory() (map[string]any, error) {
	return map[string]any{
		"newNonce": p.GetUrl() + "/nonce/new-nonce",
	}, nil
}

// GetUrl returns the URL configured for the memory provider.
func (p *MemoryDirectoryProvider) GetUrl() string {
	return p.url
}

// SetTransport is a no-op for MemoryDirectoryProvider, as it does not use a
// transport. It always returns nil.
func (p *MemoryDirectoryProvider) SetTransport(_ Transport) error {
	return nil
}

// HttpDirectoryProvider provides access to a remote directory via HTTP.
// It uses an embedded HttpTransport to make requests to the given URL.
type HttpDirectoryProvider struct {
	Url string
	*HttpTransport
}

// NewHttpDirectoryProvider creates and returns a new HttpDirectoryProvider
// with the given URL.
func NewHttpDirectoryProvider(url string) *HttpDirectoryProvider {
	return &HttpDirectoryProvider{
		Url: url,
	}
}

// Directory fetches the remote directory from the provider's URL.
// It returns a map representing the directory, or an error if the request or
// decoding fails.
func (p *HttpDirectoryProvider) Directory() (map[string]any, error) {
	r, err := p.HttpTransport.Get(p.GetUrl())
	if err != nil {
		return nil, err
	}
	dir := map[string]any{}
	err = BodyAsJson(r, &dir)
	if err != nil {
		return nil, err
	}
	return dir, nil
}

// GetUrl returns the URL configured for the HTTP directory provider.
func (p *HttpDirectoryProvider) GetUrl() string {
	return p.Url
}

// SetTransport sets the HTTP transport for the provider. It attempts to cast
// the given Transport to *HttpTransport. Returns an error if the cast fails.
func (p *HttpDirectoryProvider) SetTransport(t Transport) error {
	var ok bool
	p.HttpTransport, ok = t.(*HttpTransport)
	if !ok {
		return errors.New("was not able to cast Transport to HttpTransport")
	}
	return nil
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
