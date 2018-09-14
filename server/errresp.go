package server

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

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
func (s *Server) invalidParameterResp(c *Context, err error, msg string) {
	log.Debugf("InvalidParameterResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	InvalidParameterResp(c, msg)
}

// AuthenticationErrorResp responds a error JSON because of the authentication failed.
func AuthenticationErrorResp(c *Context, msg string) {
	if msg == "" {
		msg = "Authentication Error"
	}
	errResp(c, http.StatusUnauthorized, "AuthenticationError", msg)
}
func (s *Server) authenticationErrorResp(c *Context, err error, msg string) {
	log.Debugf("AuthenticationErrorResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	AuthenticationErrorResp(c, msg)
}

// AuthenticationExpiredResp responds a error JSON because of the authentication expired.
func AuthenticationExpiredResp(c *Context, msg string) {
	if msg == "" {
		msg = "Authentication Expired"
	}
	errResp(c, http.StatusUnauthorized, "AuthenticationExpired", msg)
}
func (s *Server) authenticationExpiredResp(c *Context, err error, msg string) {
	log.Debugf("AuthenticationExpiredResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	AuthenticationExpiredResp(c, msg)
}

// ForbiddenResp responds a error JSON because of the unauthorized operation.
func ForbiddenResp(c *Context, msg string) {
	if msg == "" {
		msg = http.StatusText(http.StatusForbidden)
	}
	errResp(c, http.StatusForbidden, "Forbidden", msg)
}
func (s *Server) forbiddenResp(c *Context, err error, msg string) {
	log.Debugf("ForbiddenResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	ForbiddenResp(c, msg)
}

// NotFoundResp responds a error JSON because resource is not found.
func NotFoundResp(c *Context, msg string) {
	if msg == "" {
		msg = http.StatusText(http.StatusNotFound)
	}
	errResp(c, http.StatusNotFound, "NotFound", msg)
}
func (s *Server) notFoundResp(c *Context, err error, msg string) {
	log.Debugf("NotFoundResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	NotFoundResp(c, msg)
}

// InternalServerErrorResp responds a error JSON because of an unexpected error.
func InternalServerErrorResp(c *Context, msg string) {
	if msg == "" {
		msg = http.StatusText(http.StatusInternalServerError)
	}
	errResp(c, http.StatusInternalServerError, "InternalServerError", msg)
}
func (s *Server) internalServerErrorResp(c *Context, err error, msg string) {
	log.Errorf("InternalServerErrorResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	InternalServerErrorResp(c, msg)
}

// TimeoutErrorResp responds a error JSON because of operation timeout.
func TimeoutErrorResp(c *Context, msg string) {
	if msg == "" {
		msg = http.StatusText(http.StatusGatewayTimeout)
	}
	errResp(c, http.StatusGatewayTimeout, "TimeoutError", msg)
}
func (s *Server) timeoutErrorResp(c *Context, err error, msg string) {
	log.Debugf("TimeoutErrorResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	TimeoutErrorResp(c, msg)
}
