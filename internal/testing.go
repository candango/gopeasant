package peasant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type HttpTestRunner struct {
	clearBody   bool
	clearHeader bool
	clearHFunc  bool
	body        io.Reader
	header      http.Header
	handler     http.Handler
	hFunc       func(http.ResponseWriter, *http.Request)
	method      string
	path        string
	t           *testing.T
}

// NewHttpTestRunner create a new runner with empty headers, method as
// http.MethodGet and root path as default.
func NewHttpTestRunner(t *testing.T) *HttpTestRunner {
	r := &HttpTestRunner{}
	r.header = http.Header{}
	r.method = http.MethodGet
	r.path = "/"
	r.t = t
	return r
}

// Clear set body and hFunc as nil, and empty the header.
// Also clearBody, clearHFunc, and clearHeader flags are set to false.
func (r *HttpTestRunner) Clear() *HttpTestRunner {
	r.body = nil
	r.clearBody = false
	r.hFunc = nil
	r.clearHFunc = false
	r.header = http.Header{}
	r.clearHeader = false
	return r
}

// ClearBodyAfter set body to be cleared after HttpTestRunner.Run execution
func (r *HttpTestRunner) ClearBodyAfter() *HttpTestRunner {
	r.clearBody = true
	return r
}

// ClearBodyAfter set hFunc to be cleared after HttpTestRunner.Run execution
func (r *HttpTestRunner) ClearFuncAfter() *HttpTestRunner {
	r.clearHFunc = true
	return r
}

// ClearHeaderAfter set the header to be cleared after HttpTestRunner.Run
// execution
func (r *HttpTestRunner) ClearHeaderAfter() *HttpTestRunner {
	r.clearHeader = true
	return r
}

// WithFunc set a function to be exectued by the runner.
//
// If a function is defined, it will bypass the handler.
//
// Use ClearFuncAfter to run the function once and clear it for the next
// HttpTestRunner.Run execution.
func (r *HttpTestRunner) WithFunc(
	hFunc func(http.ResponseWriter, *http.Request)) *HttpTestRunner {
	r.hFunc = hFunc
	return r
}

// WithHandler set a handler to be executed by the runner.
//
// If a function is defined, it will bypass this handler.
//
// Use ClearFuncAfter to run the function once and clear it for the next
// HttpTestRunner.Run execution.
func (r *HttpTestRunner) WithHandler(handler http.Handler) *HttpTestRunner {
	r.handler = handler
	return r
}

// WithHeader add a key/value pair to be added to the header
func (r *HttpTestRunner) WithHeader(key string, value string) *HttpTestRunner {
	r.header.Add(key, value)
	return r
}

// WithPath set the path to be executed by the runner
func (r *HttpTestRunner) WithPath(path string) *HttpTestRunner {
	r.path = path
	return r
}

// WithBody set HttpTestRunner.body using an io.Reader
func (r *HttpTestRunner) WithBody(body io.Reader) *HttpTestRunner {
	r.body = body
	return r
}

// WithJsonBody set HttpTestRunner.body using an interface
func (r *HttpTestRunner) WithJsonBody(typedBody any) *HttpTestRunner {
	marshaledTypedRequest, _ := json.Marshal(typedBody)
	r.WithBody(bytes.NewReader(marshaledTypedRequest))
	return r
}

// WithStringBody set HttpTestRunner.body using a string
func (r *HttpTestRunner) WithStringBody(stringBody string) *HttpTestRunner {
	r.WithBody(bytes.NewReader([]byte(stringBody)))
	return r
}

// WithMethod set the method to be used by the runner execution
func (r *HttpTestRunner) WithMethod(method string) *HttpTestRunner {
	r.method = strings.ToUpper(method)
	return r
}

// runMethod executes the http method following HttpTestRunner configuration.
// If hFunc is defined it will bypass a handler even if defined.
func (r *HttpTestRunner) runMethod() (*http.Response, error) {
	handler := r.handler
	if r.hFunc != nil {
		handler = http.HandlerFunc(r.hFunc)
	}
	s := httptest.NewServer(handler)
	defer s.Close()
	u, err := url.Parse(s.URL + r.path)
	if err != nil {
		r.t.Error(err)
		r.t.FailNow()
	}
	var req *http.Request
	req, err = http.NewRequest(r.method, u.String(), r.body)
	req.Header = r.header
	if err != nil {
		r.t.Error(err)
		r.t.FailNow()
	}
	client := &http.Client{}
	var res *http.Response
	res, err = client.Do(req)
	if err != nil {
		r.t.Error(err)
		r.t.FailNow()
	}
	return res, err
}

func (r *HttpTestRunner) reset() {
	if r.clearBody {
		r.body = nil
		r.clearBody = false
	}
	if r.clearHFunc {
		r.hFunc = nil
		r.clearHFunc = false
	}
	if r.clearHeader {
		r.header = http.Header{}
		r.clearHeader = false
	}
}

// Run the http method if it is allowed
func (r *HttpTestRunner) Run() (resp *http.Response, err error) {
	defer r.reset()
	switch r.method {
	case http.MethodDelete, http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut:
		resp, err = r.runMethod()
	default:
		resp, err = nil, errors.New(
			fmt.Sprintf("unsupported method: %s", r.method))
	}
	return resp, err
}

// resetMethod change to the previous method if it is the case
func (r *HttpTestRunner) resetMethod(previous string) {
	if previous != r.method {
		r.method = previous
	}
}

// Delete set method to http.Delete, call HttpTestRunner.Run, and reset method
// to the previous if it is the case.
func (r *HttpTestRunner) Delete() (resp *http.Response, err error) {
	previousMethod := r.method
	r.method = http.MethodDelete
	defer r.resetMethod(previousMethod)
	r.method = http.MethodDelete
	resp, err = r.Run()
	return resp, err
}

// Get set method to http.Get, call HttpTestRunner.Run, and reset method to the
// previous if it is the case.
func (r *HttpTestRunner) Get() (resp *http.Response, err error) {
	previousMethod := r.method
	r.method = http.MethodGet
	defer r.resetMethod(previousMethod)
	resp, err = r.Run()
	return resp, err
}

// Head set method to http.Head, call HttpTestRunner.Run, and reset method to
// the previous if it is the case.
func (r *HttpTestRunner) Head() (resp *http.Response, err error) {
	previousMethod := r.method
	r.method = http.MethodHead
	defer r.resetMethod(previousMethod)
	resp, err = r.Run()
	return resp, err
}

// Post set method to http.Post, call HttpTestRunner.Run, and reset method to
// the previous if it is the case.
func (r *HttpTestRunner) Post() (resp *http.Response, err error) {
	previousMethod := r.method
	r.method = http.MethodPost
	defer r.resetMethod(previousMethod)
	resp, err = r.Run()
	return resp, err
}

// Put set method to http.Put, call HttpTestRunner.Run, and reset method to
// the previous if it is the case.
func (r *HttpTestRunner) Put() (resp *http.Response, err error) {
	previousMethod := r.method
	r.method = http.MethodPut
	defer r.resetMethod(previousMethod)
	resp, err = r.Run()
	return resp, err
}

// BodyAsString returns the body of a request as string
func BodyAsString(t *testing.T, resp http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	return string(body)
}

// BodyAsJson unmarshal the body of a request to json
func BodyAsJson(t *testing.T, resp *http.Response, jsonBody any) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(b, jsonBody)
	if err != nil {
		t.Error(err)
	}
}
