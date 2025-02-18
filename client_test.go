package peasant

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestTransport struct {
	*HttpTransport
}

func NewTestTransport(tr *HttpTransport) *TestTransport {
	return &TestTransport{tr}
}

func (tt *TestTransport) Directory() (map[string]any, error) {
	d, err := tt.HttpTransport.Directory()
	if err != nil {
		return nil, err
	}
	d["doSomething"] = tt.GetUrl() + "/nonce/do-nonced-something"
	return d, nil
}

func (tt *TestTransport) DoSomething(t *testing.T) (string, error) {
	nonce, err := tt.NewNonce()
	if err != nil {
		return "", err
	}

	d, err := tt.Directory()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, d["doSomething"].(string), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("nonce", nonce)

	res, err := tt.Client.Do(req)
	if err != nil {
		return "", err
	}

	b, err := BodyAsString(res)
	if err != nil {
		return "", err
	}

	if res.StatusCode > 299 {
		return "", errors.New(res.Status)
	}
	return b, nil
}

type TestPeasant struct {
	*Peasant
}

func NewTestPesant(p *Peasant) *TestPeasant {
	return &TestPeasant{p}
}

func (p *TestPeasant) DoSomething(t *testing.T) (string, error) {
	return p.Transport.(*TestTransport).DoSomething(t)
}

func NewServer(t *testing.T) *httptest.Server {
	handler := http.NewServeMux()
	handler.Handle("/nonce/", http.StripPrefix(
		"/nonce", NewNoncedFuncServeMux(t)))
	return httptest.NewServer(handler)
}

func TestHttpTransport(t *testing.T) {
	server := NewServer(t)
	dp := &MemoryDirectoryProvider{server.URL}
	ht := NewHttpTransport(dp)

	t.Run("NewNonceUrl should return url if DirectoryKey is valid and error otherwise", func(t *testing.T) {
		directoryKey := ht.DirectoryKey
		url, err := ht.NewNonceUrl()
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, fmt.Sprintf("%s/nonce/new-nonce", server.URL), url)
		ht.DirectoryKey = "BadDirectoryKey"
		url, err = ht.NewNonceUrl()
		if err != nil {
			assert.NotNil(t, err)
		}
		assert.Empty(t, url)
		ht.DirectoryKey = directoryKey
	})

	t.Run("NewNonce should return value if NonceKey is value and error otherwise", func(t *testing.T) {
		nonceKey := ht.NonceKey
		value, err := ht.NewNonce()
		ht.NewNonce()
		if err != nil {
			t.Error(err)
		}
		assert.NotEmpty(t, value)
		ht.NonceKey = "BadNonceKey"
		value, err = ht.NewNonce()
		if err != nil {
			assert.NotNil(t, err)
		}
		assert.Empty(t, value)
		ht.NonceKey = nonceKey
	})

	t.Run("Plain Peasant and Transport New Nonce Generation", func(t *testing.T) {
		p := NewPeasant(ht)
		nonce, err := p.NewNonce()
		if err != nil {
			t.Error(err)
		}
		assert.NotNil(t, nonce)
	})

	t.Run("Custom Request OK", func(t *testing.T) {
		p := NewTestPesant(NewPeasant(NewTestTransport(ht)))
		something, err := p.DoSomething(t)
		if err != nil {
			t.Error(err)
		}
		assert.True(t, strings.HasPrefix(something, "Func done with nonce "))

	})
}
