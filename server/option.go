package server

import (
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"

	"github.com/mikunalpha/httpsrvtpl/store"
)

// Option is a func accepts Server to do configuration.
type Option func(*Server)

// OptAddress set address which sila listen on.
func OptAddress(address string) Option {
	return func(s *Server) {
		s.address = address
	}
}

// OptAutoCert gets LetsEncrypt for domains automatically. If cacheDirPath is empty, it
// will not cache any certs.
func OptAutoCert(cacheDirPath string, domains ...string) Option {
	return func(s *Server) {
		s.enableAutoCert = true
		s.autoCertCacheDirPath = cacheDirPath
		s.autoCertDomains = domains
	}
}

// OptStore assigns the implementation of store.Store to server.
func OptStore(st store.Store) Option {
	return func(s *Server) {
		s.store = st
	}
}

// optAllowMethodOverrideOnce is used to ensure that OptAllowMethodOverride is only activated once.
var optAllowMethodOverrideOnce = &sync.Once{}

// OptAllowMethodOverride allows a request override its method with header X-HTTP-Method-Override.
func OptAllowMethodOverride() Option {
	return func(s *Server) {
		optAllowMethodOverrideOnce.Do(func() {
			s.handler = func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					method := r.Header.Get("X-HTTP-Method-Override")
					if method != "" {
						r.Method = method
					}
					next.ServeHTTP(w, r)
				})
			}(s.handler)
		})
	}
}

// optAddPingHandlerOnce is used to ensure that OptAddPingHandler is only activated once.
var optAddPingHandlerOnce = &sync.Once{}

// OptAddPingHandler add [GET] /ping route into router.
func OptAddPingHandler() Option {
	return func(s *Server) {
		optAddPingHandlerOnce.Do(func() {
			s.routerEngine.GET("/ping", func(c *Context) {
				c.JSON(http.StatusOK, &struct {
					Ping string `json:"ping"`
				}{"pong"})
			})
		})
	}
}

// optAddDebugHandlerOnce is used to ensure that OptAddDebugHandler is only activated once.
var optAddDebugHandlerOnce = &sync.Once{}

// OptAddDebugHandler add below routes into router.
// [GET] /debug/pprof
// [GET] /debug/pprof/cmdline
// [GET] /debug/pprof/profile
// [GET] /debug/pprof/symbol
// [POST] /debug/pprof/symbol
// [GET] /debug/pprof/trace
// [GET] /debug/block
// [GET] /debug/goroutine
// [GET] /debug/heap
// [GET] /debug/mutex
// [GET] /debug/threadcreate
func OptAddDebugHandler() Option {
	httpToGin := func(h http.HandlerFunc) HandlerFunc {
		handler := h
		return func(c *Context) {
			handler.ServeHTTP(c.Writer, c.Request)
		}
	}
	pprofIndex := func(c *Context) {
		pprof.Handler(strings.TrimPrefix(c.Request.URL.Path, "/debug/")).ServeHTTP(c.Writer, c.Request)
	}
	return func(s *Server) {
		optAddDebugHandlerOnce.Do(func() {
			s.routerEngine.GET("/debug/pprof", httpToGin(pprof.Index))
			s.routerEngine.GET("/debug/pprof/cmdline", httpToGin(pprof.Cmdline))
			s.routerEngine.GET("/debug/pprof/profile", httpToGin(pprof.Profile))
			s.routerEngine.GET("/debug/pprof/symbol", httpToGin(pprof.Symbol))
			s.routerEngine.POST("/debug/pprof/symbol", httpToGin(pprof.Symbol))
			s.routerEngine.GET("/debug/pprof/trace", httpToGin(pprof.Trace))
			s.routerEngine.GET("/debug/block", pprofIndex)
			s.routerEngine.GET("/debug/goroutine", pprofIndex)
			s.routerEngine.GET("/debug/heap", pprofIndex)
			s.routerEngine.GET("/debug/mutex", pprofIndex)
			s.routerEngine.GET("/debug/threadcreate", pprofIndex)
		})
	}
}
