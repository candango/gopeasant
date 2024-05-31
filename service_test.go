package peasant

import (
	"math/rand"
	"net/http"
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

type InMemoryNonceService struct {
	nonceMap map[string]*interface{}
}

func (s *InMemoryNonceService) Block(resp http.ResponseWriter,
	req *http.Request) error {
	return nil
}

func (s *InMemoryNonceService) Clear(nonce string) error {
	_, ok := s.nonceMap[nonce]
	if !ok {
		return nil
	}
	delete(s.nonceMap, nonce)
	return nil
}

func (s *InMemoryNonceService) Consume(resp http.ResponseWriter,
	req *http.Request, nonce string) (bool, error) {
	_, ok := s.nonceMap[nonce]
	if !ok {
		return false, nil
	}
	err := s.Clear(nonce)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *InMemoryNonceService) GetNonce(req *http.Request) (string, error) {
	nonce := randomString(32)
	s.nonceMap[nonce] = nil
	ticker := time.NewTicker(2000 * time.Millisecond)
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

func (s *InMemoryNonceService) Provided(_ http.ResponseWriter,
	_ *http.Request) (bool, error) {
	return false, nil
}

type NoncedHandler struct {
	http.Handler
	service NonceService
}

func NewNoncedHandler() *NoncedHandler {
	h := http.NewServeMux()
	h.Handle("/new-nonce", &NoncedHandler{})
	s := &InMemoryNonceService{
		nonceMap: make(map[string]*interface{}),
	}
	return &NoncedHandler{
		Handler: h,
		service: s,
	}

}

func (h *NoncedHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	method := req.Method
	if method != http.MethodHead {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	nonce, err := h.service.GetNonce(req)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("nonce", nonce)
}

func TestServer(t *testing.T) {
	runner := testrunner.NewHttpTestRunner(t).WithHandler(NewNoncedHandler())

	t.Run("Head Request tests", func(t *testing.T) {
		t.Run("Request OK", func(t *testing.T) {
			res, err := runner.WithPath("/new-nonce").Head()
			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, "200 OK", res.Status)
			assert.Equal(t, http.NoBody, res.Body)
			assert.Equal(t, 32, len(res.Header.Get("nonce")))
			res, err = runner.WithPath("/new-nonce").Get()
			assert.Equal(t, "405 Method Not Allowed", res.Status)
			assert.Equal(t, http.NoBody, res.Body)
		})
	})
}
