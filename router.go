// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package router 0.2.8 provides fast HTTP request router.

The router matches incoming requests by the request method and the path.
If a handle is registered for this path and method, the router delegates the
request to defined handler.
The router package is useful to prepare a RESTful API for Go services.
It has JSON output, which bind automatically for relevant type of data.
The router has timer feature to display duration of request handling in the header

Simplest example (serve static route):

	package main

	import (
		"github.com/takama/router"
	)

	func Hello(c *router.Control) {
		c.Body("Hello")
	}

	func main() {
		r := router.New()
		r.GET("/hello", Hello)

		// Listen and serve on 0.0.0.0:8888
		r.Listen(":8888")
	}


Check it:

	curl -i http://localhost:8888/hello

Serve dynamic route with parameter:

	func main() {
		r := router.New()
		r.GET("/hello/:name", func(c *router.Control) {
			c.Code(200).Body("Hello " + c.Get(":name"))
		})

		// Listen and serve on 0.0.0.0:8888
		r.Listen(":8888")
	}

Checks JSON Content-Type automatically:

	func main() {
		r := router.New()
		r.GET("/settings/database/:db", func(c *router.Control) {
			data := map[string]map[string]string{
				"Database settings": {
					"database": c.Get(":db"),
					"host":     "localhost",
					"port":     "3306",
				},
			}
			c.Code(200).Body(data)
		})
		// Listen and serve on 0.0.0.0:8888
		r.Listen(":8888")
	}

Go Router
*/
package router

import (
	"log"
	"net/http"
	"strings"
)

// Router represents a multiplexer for HTTP requests.
type Router struct {
	// List of handlers which accociated with known http methods (GET, POST ...)
	handlers map[string]*parser

	// NotFound is called when unknown HTTP method or a handler not found.
	// If it is not set, http.NotFound is used.
	// Please overwrite this if need your own NotFound handler.
	NotFound Handle

	// PanicHandler is called when panic happen.
	// The handler prevents your server from crashing and should be used to return
	// http status code http.StatusInternalServerError (500)
	PanicHandler Handle

	// CustomHandler is called allways if defined
	CustomHandler func(Handle) Handle

	// Logger activates logging user function for each requests
	Logger Handle
}

// Handle type is aliased to type of handler function.
type Handle func(*Control)

// New it returns a new multiplexer (Router).
func New() *Router {
	return &Router{handlers: make(map[string]*parser)}
}

// GET is a shortcut for Router Handle("GET", path, handle)
func (r *Router) GET(path string, h Handle) {
	r.Handle("GET", path, h)
}

// POST is a shortcut for Router Handle("POST", path, handle)
func (r *Router) POST(path string, h Handle) {
	r.Handle("POST", path, h)
}

// PUT is a shortcut for Router Handle("PUT", path, handle)
func (r *Router) PUT(path string, h Handle) {
	r.Handle("PUT", path, h)
}

// DELETE is a shortcut for Router Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, h Handle) {
	r.Handle("DELETE", path, h)
}

// HEAD is a shortcut for Router Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, h Handle) {
	r.Handle("HEAD", path, h)
}

// Handle registers a new request handle with the given path and method.
func (r *Router) Handle(method, path string, h Handle) {
	if r.handlers[method] == nil {
		r.handlers[method] = newParser()
	}
	r.handlers[method].register(path, h)
}

// Handler allows the usage of an http.Handler as a request handle.
func (r *Router) Handler(method, path string, handler http.Handler) {
	r.Handle(method, path,
		func(c *Control) {
			handler.ServeHTTP(c.Writer, c.Request)
		},
	)
}

// HandlerFunc allows the usage of an http.HandlerFunc as a request handle.
func (r *Router) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Handle(method, path,
		func(c *Control) {
			handler(c.Writer, c.Request)
		},
	)
}

// Lookup returns handler and URL parameters that associated with path.
func (r *Router) Lookup(method, path string) (Handle, []Param, bool) {
	if parser := r.handlers[method]; parser != nil {
		return parser.get(path)
	}
	return nil, nil, false
}

// Listen and serve on requested host and port.
func (r *Router) Listen(hostPort string) {
	if err := http.ListenAndServe(hostPort, r); err != nil {
		log.Fatal(err)
	}
}

// ServeHTTP implements http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if recovery := recover(); recovery != nil {
			if r.PanicHandler != nil {
				c := &Control{Request: req, Writer: w}
				r.PanicHandler(c)
			} else {
				log.Println("Recovered in handler:", req.Method, req.URL.Path)
			}
		}
	}()
	if r.Logger != nil {
		c := &Control{Request: req, Writer: w}
		r.Logger(c)
	}
	if _, ok := r.handlers[req.Method]; ok {
		if handle, params, ok := r.handlers[req.Method].get(req.URL.Path); ok {
			c := &Control{Request: req, Writer: w}
			if len(params) > 0 {
				c.Params = append(c.Params, params...)
			}
			if r.CustomHandler != nil {
				r.CustomHandler(handle)(c)
			} else {
				handle(c)
			}
			return
		}
	}
	allowed := make([]string, 0, len(r.handlers))
	for method, parser := range r.handlers {
		if _, _, ok := parser.get(req.URL.Path); ok {
			allowed = append(allowed, method)
		}
	}

	if len(allowed) == 0 {
		if r.NotFound != nil {
			c := &Control{Request: req, Writer: w}
			r.NotFound(c)
		} else {
			http.NotFound(w, req)
		}
		return
	}

	w.Header().Add("Allow", strings.Join(allowed, ", "))
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
