package peasant

import (
	"net/http"
	"testing"

	"github.com/candango/gopeasant/internal/testrunner"
	"github.com/stretchr/testify/assert"
)

func hFunc(res http.ResponseWriter, req *http.Request) {
	response := "A test"
	_, err := res.Write([]byte(response))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func TestWithHandler(t *testing.T) {
	runner := testrunner.NewHttpTestRunner(t)

	t.Run("With func, clear func after", func(t *testing.T) {
		res, err := runner.WithFunc(hFunc).ClearFuncAfter().Get()
		if err != nil {
			t.Error(err)
		}
		response := testrunner.BodyAsString(t, res)
		assert.Equal(t, "200 OK", res.Status)
		assert.Equal(t, response, "A test")
	})
}
