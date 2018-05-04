package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.PanicLevel)
}

func TestServer(t *testing.T) {
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

	s := New("0.0.0.0:8888")

	// Test addRoutes
	s.addRoutes("/api/v1", nil, []route{
		{"GET", "/hello", func(c *Context) { c.String(http.StatusOK, "hello") }},
		{"POST", "/hello", func(c *Context) { c.String(http.StatusOK, "hello") }},
		{"PATCH", "/hello", func(c *Context) { c.String(http.StatusOK, "hello") }},
		{"PUT", "/hello", func(c *Context) { c.String(http.StatusOK, "hello") }},
		{"DELETE", "/hello", func(c *Context) { c.String(http.StatusOK, "hello") }},
		{"OPTIONS", "/hello", func(c *Context) { c.String(http.StatusOK, "hello") }},
		{"ANY", "/helloany", func(c *Context) { c.String(http.StatusOK, "hello") }},
	})
	status, respBody := sendRequestFunc(s.handler, "GET", "/api/v1/hello", nil, nil)
	assert.Equal(t, http.StatusOK, status, "should be equal")
	assert.Equal(t, "hello", respBody, "should be equal")

	// Test recover middleware
	s.addRoutes("", nil, []route{
		{"GET", "/panic", func(c *Context) { panic("panic") }},
	})
	status, respBody = sendRequestFunc(s.handler, "GET", "/panic", nil, nil)
	assert.Equal(t, http.StatusInternalServerError, status, "should be equal")
	assert.Equal(t, `{"error":{"code":"InternalServerError","msg":"Internal Server Error"}}`, respBody, "should be equal")

	// Test errresp uitl func
	type testCase struct {
		Method             string
		Path               string
		Headers            H
		RequestBody        io.Reader
		Handler            HandlerFunc
		ExpectedStatusCode int
		ExpectedBodyString string
	}
	testCases := []testCase{
		{"GET", "/invalidparameter", nil, nil, func(c *Context) { s.invalidParameterResp(c, nil, "") }, http.StatusBadRequest, `{"error":{"code":"InvalidParameter","msg":"Invalid Parameter"}}`},
		{"GET", "/authenticationerror", nil, nil, func(c *Context) { s.authenticationErrorResp(c, nil, "") }, http.StatusUnauthorized, `{"error":{"code":"AuthenticationError","msg":"Authentication Error"}}`},
		{"GET", "/authenticationexpired", nil, nil, func(c *Context) { s.authenticationExpiredResp(c, nil, "") }, http.StatusUnauthorized, `{"error":{"code":"AuthenticationExpired","msg":"Authentication Expired"}}`},
		{"GET", "/forbidden", nil, nil, func(c *Context) { s.forbiddenResp(c, nil, "") }, http.StatusForbidden, `{"error":{"code":"Forbidden","msg":"Forbidden"}}`},
		{"GET", "/notfound", nil, nil, func(c *Context) { s.notFoundResp(c, nil, "") }, http.StatusNotFound, `{"error":{"code":"NotFound","msg":"Not Found"}}`},
		{"GET", "/internalservererror", nil, nil, func(c *Context) { s.internalServerErrorResp(c, nil, "") }, http.StatusInternalServerError, `{"error":{"code":"InternalServerError","msg":"Internal Server Error"}}`},
		{"GET", "/timeouterror", nil, nil, func(c *Context) { s.timeoutErrorResp(c, nil, "") }, http.StatusGatewayTimeout, `{"error":{"code":"TimeoutError","msg":"Gateway Timeout"}}`},
	}
	for i := range testCases {
		switch testCases[i].Method {
		case "GET":
			s.routerEngine.GET(testCases[i].Path, testCases[i].Handler)
		case "POST":
			s.routerEngine.POST(testCases[i].Path, testCases[i].Handler)
		case "PATCH":
			s.routerEngine.PATCH(testCases[i].Path, testCases[i].Handler)
		case "PUT":
			s.routerEngine.PUT(testCases[i].Path, testCases[i].Handler)
		case "DELETE":
			s.routerEngine.DELETE(testCases[i].Path, testCases[i].Handler)
		case "OPTIONS":
			s.routerEngine.OPTIONS(testCases[i].Path, testCases[i].Handler)
		case "ANY":
			s.routerEngine.Any(testCases[i].Path, testCases[i].Handler)
		}
	}
	for _, c := range testCases {
		status, respBody = sendRequestFunc(s.handler, c.Method, c.Path, c.Headers, c.RequestBody)
		assert.Equal(t, c.ExpectedStatusCode, status, "should be equal")
		assert.Equal(t, c.ExpectedBodyString, respBody, "should be equal")
	}

	// Test run
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	time.AfterFunc(3*time.Second, func() {
		stop <- os.Interrupt
	})
	s.Run()
	<-stop
	// Use sync.Mutex to avoid data race on s.running
	s.mu.Lock()
	assert.Equal(t, true, s.running, "should be equal")
	s.mu.Unlock()
	s.Stop()
	s.mu.Lock()
	assert.Equal(t, false, s.running, "should be equal")
	s.mu.Unlock()
	s.Stop()
}
