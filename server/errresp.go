package server

import "net/http"

// errResp responds a error JSON.
func errResp(c *Context, status int, code, msg string) {
	ers := &struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"msg"`
		} `json:"error"`
	}{}
	ers.Error.Code = code
	ers.Error.Message = msg
	c.JSON(status, ers)
}

// InvalidParameterResp responds a error JSON because of the invalid parameter.
func InvalidParameterResp(c *Context, msg string) {
	if msg == "" {
		msg = "Invalid Parameter"
	}
	errResp(c, http.StatusBadRequest, "InvalidParameter", msg)
}

// AuthenticationErrorResp responds a error JSON because of the authentication failed.
func AuthenticationErrorResp(c *Context, msg string) {
	if msg == "" {
		msg = "Authentication Error"
	}
	errResp(c, http.StatusUnauthorized, "AuthenticationError", msg)
}

// AuthenticationExpiredResp responds a error JSON because of the authentication expired.
func AuthenticationExpiredResp(c *Context, msg string) {
	if msg == "" {
		msg = "Authentication Expired"
	}
	errResp(c, http.StatusUnauthorized, "AuthenticationExpired", msg)
}

// ForbiddenResp responds a error JSON because of the unauthorized operation.
func ForbiddenResp(c *Context, msg string) {
	if msg == "" {
		msg = http.StatusText(http.StatusForbidden)
	}
	errResp(c, http.StatusForbidden, "Forbidden", msg)
}

// NotFoundResp responds a error JSON because resource is not found.
func NotFoundResp(c *Context, msg string) {
	if msg == "" {
		msg = http.StatusText(http.StatusNotFound)
	}
	errResp(c, http.StatusNotFound, "NotFound", msg)
}

// InternalServerErrorResp responds a error JSON because of an unexpected error.
func InternalServerErrorResp(c *Context, msg string) {
	if msg == "" {
		msg = http.StatusText(http.StatusInternalServerError)
	}
	errResp(c, http.StatusInternalServerError, "InternalServerError", msg)
}

// TimeoutErrorResp responds a error JSON because of operation timeout.
func TimeoutErrorResp(c *Context, msg string) {
	if msg == "" {
		msg = http.StatusText(http.StatusGatewayTimeout)
	}
	errResp(c, http.StatusGatewayTimeout, "TimeoutError", msg)
}
