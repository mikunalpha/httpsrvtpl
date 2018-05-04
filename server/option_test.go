package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mikunalpha/httpsrvtpl/store/mock"
	"github.com/stretchr/testify/assert"
)

func TestServerOptions(t *testing.T) {
	type H map[string]string

	var sendRequestFunc = func(handler http.Handler, method, path string, headers map[string]string, requestBody io.Reader) (int, string) {
		req := httptest.NewRequest(method, "http://xxx.com"+path, requestBody)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		respRecorder := httptest.NewRecorder()
		handler.ServeHTTP(respRecorder, req)
		return respRecorder.Code, respRecorder.Body.String()
	}

	st := mock.New()

	opts := []Option{
		OptAddress("0.0.0.0:9876"),
		OptAutoCert("./ssl", "abc.fake.com"),
		OptStore(st),
		OptAddPingHandler(),
		OptAddDebugHandler(),
		OptAllowMethodOverride(),
	}

	s := New("0.0.0.0:80", opts...)

	// Test OptAddress
	assert.Equal(t, "0.0.0.0:9876", s.address, "should be equal")

	// Test OptAutoCert
	assert.Equal(t, true, s.enableAutoCert, "should be equal")
	assert.Equal(t, "./ssl", s.autoCertCacheDirPath, "should be equal")
	assert.Equal(t, []string{"abc.fake.com"}, s.autoCertDomains, "should be equal")

	// Test OptStore
	assert.Equal(t, st, s.store, "should be equal")

	// Test OptAddPingHandler
	status, respBody := sendRequestFunc(s.handler, "GET", "/ping", nil, nil)
	assert.Equal(t, http.StatusOK, status, "should be equal")
	assert.Equal(t, `{"ping":"pong"}`, respBody, "should be equal")

	// Test OptAddDebugHandler
	status, respBody = sendRequestFunc(s.handler, "GET", "/debug/pprof", nil, nil)
	assert.Equal(t, http.StatusOK, status, "should be equal")
	status, respBody = sendRequestFunc(s.handler, "GET", "/debug/goroutine", nil, nil)
	assert.Equal(t, http.StatusOK, status, "should be equal")

	// Test OptAllowMethodOverride
	status, respBody = sendRequestFunc(s.handler, "GET", "/debug/pprof/symbol", H{"X-HTTP-Method-Override": "PATCH"}, nil)
	assert.Equal(t, http.StatusNotFound, status, "should be equal")

	// Test others below ...
}
