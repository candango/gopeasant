package peasant

import (
	"net/http"
	"testing"
	"time"

	"github.com/candango/gopeasant/testrunner"
	"github.com/stretchr/testify/assert"
)

func NewNoncedServeMux(t *testing.T) http.Handler {
	s := &InMemoryNonceService{
		nonceMap: make(map[string]*interface{}),
		t:        t,
	}
	nonced := &NoncedHandler{
		service: s,
	}
	h := http.NewServeMux()
	h.HandleFunc("/new-nonce", nonced.getNonce)
	h.HandleFunc("/do-nonced-something", nonced.doNoncedFunc)
	return Nonced(h, s)
}

func TestNoncedServer(t *testing.T) {
	handler := NewNoncedServeMux(t)
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
