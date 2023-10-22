package peasant

import (
	"log"
	"net/http"
	"testing"

	"github.com/candango/gopeasant/testrunner"
	"github.com/stretchr/testify/assert"
)

type InMemoryNonceService struct {
	nonceMap map[string]*interface{}
}

func (s *InMemoryNonceService) Block(resp http.ResponseWriter,
	req *http.Request) error {
	return nil
}
func (s *InMemoryNonceService) Clear(nonce string) error {
	return nil
}
func (s *InMemoryNonceService) Consume(resp http.ResponseWriter,
	req *http.Request, nonce string) (bool, error) {
	return false, nil
}
func (s *InMemoryNonceService) GetNonce(req *http.Request) (string, error) {
	return "", nil
}
func (s *InMemoryNonceService) provided(req http.ResponseWriter,
	res *http.Request) (bool, error) {
	return false, nil
}

type NoncedHandler struct {
	http.Handler
	log.Logger
	service NonceService
}

func NewNoncedHandler() *NoncedHandler {
	h := http.NewServeMux()
	h.Handle("/new-nonce", &NonceHandler{})
	return &NoncedHandler{
		Handler: h,
	}

}

type NonceHandler struct {
	service NonceService
}

func (h *NonceHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	method := req.Method
	if method != http.MethodHead {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	res.Header().Add("return", "OK")
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
			assert.Equal(t, "OK", res.Header.Get("return"))
		})
	})
}
