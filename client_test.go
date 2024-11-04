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
	"errors"
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

func (tt *TestTransport) Directory() (map[string]interface{}, error) {
	d, err := tt.HttpTransport.Directory()
	if err != nil {
		return nil, err
	}
	d["doSomething"] = tt.Url + "/nonce/do-nonced-something"
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
	ht := NewHttpTransport(server.URL, "Nonce")

	t.Run("Plain Peasant and Transport", func(t *testing.T) {
		p := NewPeasant(ht)
		nonce, err := p.NewNonce()
		if err != nil {
			t.Error(err)
		}
		assert.NotNil(t, nonce)
	})

	t.Run("Request OK", func(t *testing.T) {
		p := NewTestPesant(NewPeasant(NewTestTransport(ht)))
		something, err := p.DoSomething(t)
		if err != nil {
			t.Error(err)
		}
		assert.True(t, strings.HasPrefix(something, "Func done with nonce "))

	})
}
