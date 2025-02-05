package dummy

import (
	"math/rand"
	"net/http"
	"strings"
	"time"
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

// DummyInMemoryNonceService implements the NonceService interface for managing
// nonces in an in-memory map.
type DummyInMemoryNonceService struct {
	nonceMap map[string]*any
}

func (s *DummyInMemoryNonceService) Block(resp http.ResponseWriter,
	req *http.Request) error {
	return nil
}

// Clear clears the nonce associated with the specified key in the in-memory
// map.
func (s *DummyInMemoryNonceService) Clear(nonce string) error {
	_, ok := s.nonceMap[nonce]
	if !ok {
		return nil
	}
	delete(s.nonceMap, nonce)
	return nil
}

// Consume consumes the nonce associated with a specified key and returns
// whether the nonce was successfully consumed and any error that occurred.
func (s *DummyInMemoryNonceService) Consume(res http.ResponseWriter,
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

func (s *DummyInMemoryNonceService) GetNonce(req *http.Request) (string, error) {
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

func (s *DummyInMemoryNonceService) Skip(r *http.Request) bool {
	if strings.Contains(r.URL.String(), "new-nonce") {
		return true
	}
	return false
}

func (s *DummyInMemoryNonceService) Provided(w http.ResponseWriter,
	r *http.Request) error {
	nonce := r.Header.Get("nonce")
	if nonce == "" {
		w.WriteHeader(http.StatusForbidden)
		return nil
	}
	return nil
}

func NewDummyInMemoryNonceService() *DummyInMemoryNonceService {
	return &DummyInMemoryNonceService{
		nonceMap: make(map[string]*any),
	}
}
