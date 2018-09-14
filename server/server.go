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

var errf = fmt.Errorf

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
		s.notFoundResp(c, errf("no route"), "")
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

	mu      sync.Mutex
	running bool
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
		defer func() {
			panic := recover()
			if panic != nil {
				log.Error("↧↧↧↧↧↧ PANIC ↧↧↧↧↧↧")
				log.Error(panic)
				for i := 3; ; i++ {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Errorf("%s:%d", file, line)
				}
				log.Error("↥↥↥↥↥↥ PANIC ↥↥↥↥↥↥")
			}
		}()

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
		defer func() {
			panic := recover()
			if panic != nil {
				log.Error("↧↧↧↧↧↧ PANIC ↧↧↧↧↧↧")
				log.Error(panic)
				for i := 3; ; i++ {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Errorf("%s:%d", file, line)
				}
				log.Error("↥↥↥↥↥↥ PANIC ↥↥↥↥↥↥")
			}
		}()

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
	s.mu.Lock()
	if !s.running {
		log.Debug("stop server: server is not running")
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Infof("stop server")
	return s.server.Shutdown(ctx)
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

				s.internalServerErrorResp(c, errf("panic when deal with request [%s] %s", c.Request.Method, c.Request.URL), "")
			}
		}()

		c.Next()
	}
}
