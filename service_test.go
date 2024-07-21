package peasant

import (
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/candango/gopeasant/testrunner"
	"github.com/stretchr/testify/assert"
)

func randomString(s int) string {
	asciiLower := "abcdefghijklmnopqrstuvwxyz"
	asciiUpper := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "012345679"
	chars := []rune(asciiLower + asciiUpper + digits)
	r := make([]rune, s)
	for i := range r {
		r[i] = chars[rand.Intn(len(chars))]
	}
	return string(r)
}

// InMemoryNonceService implements the NonceService interface for managing
// nonces in an in-memory map.
type InMemoryNonceService struct {
	nonceMap map[string]*interface{}
	t        *testing.T
}

func (s *InMemoryNonceService) Block(resp http.ResponseWriter,
	req *http.Request) error {
	return nil
}

// Clear clears the nonce associated with the specified key in the in-memory
// map.
func (s *InMemoryNonceService) Clear(nonce string) error {
	_, ok := s.nonceMap[nonce]
	if !ok {
		return nil
	}
	delete(s.nonceMap, nonce)
	return nil
}

// Consume consumes the nonce associated with a specified key and returns
// whether the nonce was successfully consumed and any error that occurred.
func (s *InMemoryNonceService) Consume(res http.ResponseWriter,
	req *http.Request) error {
	nonce := req.Header.Get("nonce")
	if nonce == "" {
		res.WriteHeader(http.StatusForbidden)
		return nil
	}
	_, ok := s.nonceMap[nonce]
	if !ok {
		res.WriteHeader(http.StatusForbidden)
		return nil
	}
	err := s.Clear(nonce)
	if err != nil {
		return err
	}
	return nil
}

func (s *InMemoryNonceService) GetNonce(req *http.Request) (string, error) {
	nonce := randomString(32)
	s.nonceMap[nonce] = nil
	ticker := time.NewTicker(250 * time.Millisecond)
	done := make(chan bool)

	go func(nonce string) {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				s.Clear(nonce)
				done <- true
			}
		}
	}(nonce)

	return nonce, nil
}

func (s *InMemoryNonceService) IsNonced(r *http.Request) bool {
	if strings.Contains(r.URL.String(), "new-nonce") {
		return false
	}
	return true
}

func (s *InMemoryNonceService) IsProvided(w http.ResponseWriter,
	r *http.Request) error {
	nonce := r.Header.Get("nonce")
	if nonce == "" {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}
	return nil
}

type NoncedHandler struct {
	http.Handler
	service NonceService
}

func (h *NoncedHandler) getNonce(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	nonce, err := h.service.GetNonce(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("nonce", nonce)
}

func (h *NoncedHandler) doNoncedFunc(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	if method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	nonce := r.Header.Get("nonce")
	w.Write([]byte("Func done with nonce " + nonce))
}

func NewNoncedFuncServeMux(t *testing.T) *http.ServeMux {
	s := &InMemoryNonceService{
		nonceMap: make(map[string]*interface{}),
		t:        t,
	}
	nonced := &NoncedHandler{
		service: s,
	}
	h := http.NewServeMux()
	h.HandleFunc("/new-nonce", NoncedHandlerFunc(s, nonced.getNonce))
	h.HandleFunc("/do-nonced-something",
		NoncedHandlerFunc(s, nonced.doNoncedFunc))
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
