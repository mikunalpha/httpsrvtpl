package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/mikunalpha/httpsrvtpl/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

// Use jsoniter as default json package.
var json = jsoniter.ConfigCompatibleWithStandardLibrary

type (
	// HandlerFunc is alias of gin.HandlerFunc.
	HandlerFunc = gin.HandlerFunc
	// Context is alias of gin.Context.
	Context = gin.Context
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableBindValidation()
}

// New accepts a address and some opts, then it returns a new Server.
func New(address string, opts ...Option) *Server {
	s := &Server{
		address:      address,
		routerEngine: gin.New(),
	}
	s.handler = s.routerEngine

	// Set the RouterEngine to use defaut recover
	s.routerEngine.NoRoute(func(c *Context) {
		s.notFoundResp(c, fmt.Errorf("request not found [%s] %s", c.Request.Method, c.Request.URL), "")
	})
	s.routerEngine.Use(s.recover())

	// Set up the opts
	for _, opt := range opts {
		opt(s)
	}

	// Set routes
	s.setRoutes()

	return s
}

// Server is responsible for supplying HTTP service.
type Server struct {
	mu      sync.Mutex
	running bool

	address      string
	routerEngine *gin.Engine
	handler      http.Handler
	server       *http.Server

	enableAutoCert       bool
	autoCertCacheDirPath string
	autoCertDomains      []string

	store store.Store

	hasAllowMethodOverride bool
	hasPingHandler         bool
	hasDebugHandler        bool
}

// recover is the default middleware used to deal with panic.
func (s *Server) recover() HandlerFunc {
	return func(c *Context) {
		defer func() {
			panic := recover()
			if panic != nil {
				log.Debug("↧↧↧↧↧↧ PANIC ↧↧↧↧↧↧")
				log.Debug(panic)
				for i := 3; ; i++ {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Debugf("%s:%d", file, line)
				}
				log.Debug("↥↥↥↥↥↥ PANIC ↥↥↥↥↥↥")

				s.internalServerErrorResp(c, fmt.Errorf("panic when deal with request [%s] %s", c.Request.Method, c.Request.URL), "")
			}
		}()

		c.Next()
	}
}

func (s *Server) invalidParameterResp(c *Context, err error, msg string) {
	log.Debugf("InvalidParameterResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	InvalidParameterResp(c, msg)
}

func (s *Server) authenticationErrorResp(c *Context, err error, msg string) {
	log.Debugf("AuthenticationErrorResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	AuthenticationErrorResp(c, msg)
}

func (s *Server) authenticationExpiredResp(c *Context, err error, msg string) {
	log.Debugf("AuthenticationExpiredResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	AuthenticationExpiredResp(c, msg)
}

func (s *Server) forbiddenResp(c *Context, err error, msg string) {
	log.Debugf("ForbiddenResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	ForbiddenResp(c, msg)
}

func (s *Server) notFoundResp(c *Context, err error, msg string) {
	log.Debugf("NotFoundResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	NotFoundResp(c, msg)
}

func (s *Server) internalServerErrorResp(c *Context, err error, msg string) {
	log.Errorf("InternalServerErrorResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	InternalServerErrorResp(c, msg)
}

func (s *Server) timeoutErrorResp(c *Context, err error, msg string) {
	log.Debugf("TimeoutErrorResp: %v from %s request [%s] %s", err, c.ClientIP(), c.Request.Method, c.Request.URL)
	TimeoutErrorResp(c, msg)
}

// route contains method, path and handlerFunc to deal with requests.
type route struct {
	method      string
	path        string
	handlerFunc HandlerFunc
}

// addRoutes adds routes into router.
func (s *Server) addRoutes(prefix string, middlewares []HandlerFunc, routes []route) {
	routesGroup := s.routerEngine.Group(prefix, middlewares...)
	for i := range routes {
		switch routes[i].method {
		case "GET":
			routesGroup.GET(routes[i].path, routes[i].handlerFunc)
		case "POST":
			routesGroup.POST(routes[i].path, routes[i].handlerFunc)
		case "PATCH":
			routesGroup.PATCH(routes[i].path, routes[i].handlerFunc)
		case "PUT":
			routesGroup.PUT(routes[i].path, routes[i].handlerFunc)
		case "DELETE":
			routesGroup.DELETE(routes[i].path, routes[i].handlerFunc)
		case "OPTIONS":
			routesGroup.OPTIONS(routes[i].path, routes[i].handlerFunc)
		case "ANY":
			routesGroup.Any(routes[i].path, routes[i].handlerFunc)
		}
	}
}

// Run starts the server to service http request.
func (s *Server) Run() {
	if s.enableAutoCert {
		s.runWithAutoTLS()
		return
	}
	s.run()
}

// run starts the server without encryption.
func (s *Server) run() {
	s.server = &http.Server{
		Addr:         s.address,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      s.handler,
	}

	go func() {
		s.mu.Lock()
		s.running = true
		s.mu.Unlock()

		log.Debugf("listen on %s", s.address)
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("listen error %s", err)
		}

		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()
}

// runWithAutoTLS starts the server without Let's Encrypt.
func (s *Server) runWithAutoTLS() {
	if len(s.autoCertDomains) == 0 {
		log.Debug("runWithAutoTLS failed: no any domain for autocert")
		return
	}
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.autoCertDomains...),
	}
	if s.autoCertCacheDirPath != "" {
		m.Cache = autocert.DirCache(s.autoCertCacheDirPath)
	}
	s.server = &http.Server{
		Addr:         s.address,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      s.handler,
		TLSConfig:    &tls.Config{GetCertificate: m.GetCertificate},
	}

	go func() {
		s.mu.Lock()
		s.running = true
		s.mu.Unlock()

		log.Infof("listen on %s", s.address)
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("listen error %s", err)
		}

		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()
}

// Stop will stop the http server service. Return error if it occurs.
func (s *Server) Stop() error {
	if !s.running {
		log.Debug("stop server: server is not running")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	log.Debug("stop server")
	return s.server.Shutdown(ctx)
}
