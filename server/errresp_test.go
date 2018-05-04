package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrResps(t *testing.T) {
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

	// Test errresp
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
		{"GET", "/invalidparameter", nil, nil, func(c *Context) { InvalidParameterResp(c, "") }, http.StatusBadRequest, `{"error":{"code":"InvalidParameter","msg":"Invalid Parameter"}}`},
		{"GET", "/authenticationerror", nil, nil, func(c *Context) { AuthenticationErrorResp(c, "") }, http.StatusUnauthorized, `{"error":{"code":"AuthenticationError","msg":"Authentication Error"}}`},
		{"GET", "/authenticationexpired", nil, nil, func(c *Context) { AuthenticationExpiredResp(c, "") }, http.StatusUnauthorized, `{"error":{"code":"AuthenticationExpired","msg":"Authentication Expired"}}`},
		{"GET", "/forbidden", nil, nil, func(c *Context) { ForbiddenResp(c, "") }, http.StatusForbidden, `{"error":{"code":"Forbidden","msg":"Forbidden"}}`},
		{"GET", "/notfound", nil, nil, func(c *Context) { NotFoundResp(c, "") }, http.StatusNotFound, `{"error":{"code":"NotFound","msg":"Not Found"}}`},
		{"GET", "/internalservererror", nil, nil, func(c *Context) { InternalServerErrorResp(c, "") }, http.StatusInternalServerError, `{"error":{"code":"InternalServerError","msg":"Internal Server Error"}}`},
		{"GET", "/timeouterror", nil, nil, func(c *Context) { TimeoutErrorResp(c, "") }, http.StatusGatewayTimeout, `{"error":{"code":"TimeoutError","msg":"Gateway Timeout"}}`},
	}
	h := gin.New()
	for i := range testCases {
		switch testCases[i].Method {
		case "GET":
			h.GET(testCases[i].Path, testCases[i].Handler)
		case "POST":
			h.POST(testCases[i].Path, testCases[i].Handler)
		case "PATCH":
			h.PATCH(testCases[i].Path, testCases[i].Handler)
		case "PUT":
			h.PUT(testCases[i].Path, testCases[i].Handler)
		case "DELETE":
			h.DELETE(testCases[i].Path, testCases[i].Handler)
		case "OPTIONS":
			h.OPTIONS(testCases[i].Path, testCases[i].Handler)
		case "ANY":
			h.Any(testCases[i].Path, testCases[i].Handler)
		}
	}
	for _, c := range testCases {
		status, respBody := sendRequestFunc(h, c.Method, c.Path, c.Headers, c.RequestBody)
		assert.Equal(t, c.ExpectedStatusCode, status, "should be equal")
		assert.Equal(t, c.ExpectedBodyString, respBody, "should be equal")
	}
}
