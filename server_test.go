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
	"testing"
	"time"

	"github.com/candango/gopeasant/dummy"
	"github.com/candango/gopeasant/testrunner"
	"github.com/stretchr/testify/assert"
)

type NoncedHandler struct {
	http.Handler
	s NonceService
}

func (h *NoncedHandler) GetNonce(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	nonce, err := h.s.GetNonce(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("nonce", nonce)
}

func (h *NoncedHandler) DoNoncedFunc(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	nonce := r.Header.Get("nonce")
	w.Write([]byte("Func done with nonce " + nonce))
}

func NewNoncedHandler(s NonceService) *NoncedHandler {
	return &NoncedHandler{
		s: s,
	}
}

func NewNoncedFuncServeMux(t *testing.T) *http.ServeMux {
	s := dummy.NewDummyInMemoryNonceService()
	nonced := NewNoncedHandler(s)
	h := http.NewServeMux()
	h.HandleFunc("/new-nonce", NoncedHandlerFunc(s, nonced.GetNonce))
	h.HandleFunc("/do-nonced-something",
		NoncedHandlerFunc(s, nonced.DoNoncedFunc))
	return h
}

func TestNoncedFuncServer(t *testing.T) {
	handler := NewNoncedFuncServeMux(t)
	runner := testrunner.NewHttpTestRunner(t).WithHandler(handler)

	t.Run("Retrieve a new nonce", func(t *testing.T) {
		t.Run("Request OK", func(t *testing.T) {
			res, err := runner.WithPath("/new-nonce").Head()
			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, "200 OK", res.Status)
			assert.Equal(t, http.NoBody, res.Body)
			assert.Equal(t, 32, len(res.Header.Get("nonce")))
		})
	})

	t.Run("Run a nonced function", func(t *testing.T) {
		t.Run("Request OK", func(t *testing.T) {
			res, err := runner.WithPath("/new-nonce").Head()
			if err != nil {
				t.Error(err)
			}
			nonce := res.Header.Get("nonce")

			res, err = runner.WithPath(
				"/do-nonced-something").WithHeader("nonce", nonce).Get()
			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, "200 OK", res.Status)
			assert.Equal(t, "Func done with nonce "+nonce,
				testrunner.BodyAsString(t, res))
			assert.Equal(t, 32, len(res.Header.Get("nonce")))
		})
		t.Run("Expired nonce", func(t *testing.T) {
			res, err := runner.WithPath("/new-nonce").Head()
			if err != nil {
				t.Error(err)
			}
			nonce := res.Header.Get("nonce")

			time.Sleep(250 * time.Millisecond)
			res, err = runner.WithPath(
				"/do-nonced-something").WithHeader("nonce", nonce).Get()
			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, "403 Forbidden", res.Status)
		})
	})
}
