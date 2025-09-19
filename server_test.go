package peasant

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/candango/gopeasant/dummy"
	"github.com/candango/httpok/testrunner"
	"github.com/stretchr/testify/assert"
)

type NoncedHandler struct {
	http.Handler
	s NonceService
}

func (h *NoncedHandler) GetNonce(w http.ResponseWriter, r *http.Request) {
	s, ok := r.Context().Value("nonce-service").(NonceService)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	method := r.Method
	if method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	nonce, err := s.GetNonce(r)
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
	return &NoncedHandler{}
}

func NewNoncedFuncServeMux(t *testing.T) http.Handler {
	s := dummy.NewDummyInMemoryNonceService()
	nonced := NewNoncedHandler(s)
	h := http.NewServeMux()
	h.HandleFunc("/directory", GetDirectory)
	h.HandleFunc("/new-nonce", NoncedHandlerFunc(s, nonced.GetNonce))
	h.HandleFunc("/do-nonced-something",
		NoncedHandlerFunc(s, nonced.DoNoncedFunc))
	return NonceServed(s, "")(h)
}

func GetDirectory(w http.ResponseWriter, r *http.Request) {
	directory := map[string]any{
		"new-nonce":    "/new-nonce",
		"do-something": "/do-nonced-something",
	}
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(directory)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
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
