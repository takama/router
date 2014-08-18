// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"encoding/json"
	"net/http"
)

const (
	MIMEJSON = "application/json"
	MIMETEXT = "text/plain"
)

// Control allows us to pass variables between middleware,
// assign Http codes and render a Body.
type Control struct {
	Request *http.Request
	Writer  http.ResponseWriter
	code    int
	Params  []Param
}

// Param is a URL parameter which represents as key and value.
type Param struct {
	Key   string
	Value string
}

// Get returns the first value associated with the given name.
// If there are no values associated with the key, an empty string is returned.
func (c *Control) Get(name string) string {
	for idx := range c.Params {
		if c.Params[idx].Key == name {
			return c.Params[idx].Value
		}
	}

	return c.Request.URL.Query().Get(name)
}

// Set adds new parameters which represents as set of key/value.
func (c *Control) Set(params []Param) {
	c.Params = append(c.Params, params...)
}

// Code assigns http status code, which returns on http request
func (c *Control) Code(code int) *Control {
	if code >= 200 && code < 600 {
		c.code = code
	}
	return c
}

// Body renders the given data into the response body
func (c *Control) Body(data interface{}) {
	var content []byte
	if str, ok := data.(string); ok {
		c.Writer.Header().Set("Content-type", MIMETEXT)
		content = []byte(str)
	} else {
		jsn, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		c.Writer.Header().Set("Content-type", MIMEJSON)
		content = jsn
	}
	if c.code > 0 {
		c.Writer.WriteHeader(c.code)
	}
	c.Writer.Write(content)
}
