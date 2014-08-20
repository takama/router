// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	// MIMEJSON - "Content-type" for JSON
	MIMEJSON = "application/json"
	// MIMETEXT - "Content-type" for TEXT
	MIMETEXT = "text/plain"
)

// Control allows us to pass variables between middleware,
// assign Http codes and render a Body.
type Control struct {

	// Request is an adapter which allows the usage of a http.Request as standard request
	Request *http.Request

	// Writer is an adapter which allows the usage of a http.ResponseWriter as standard writer
	Writer http.ResponseWriter

	// Code of HTTP status
	code int

	// Params is set of parameters
	Params []Param

	// timer used to calculate a elapsed time for handler and writing it in a response
	timer time.Time
}

// Param is a URL parameter which represents as key and value.
type Param struct {
	Key   string
	Value string
}

// Header is used to prepare a JSON header with duration triggered by UserTimer() method
type Header struct {
	Took time.Duration `json:"took"`
	Data interface{}   `json:"data"`
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

// UseTimer allow caalculate elapsed time of request handling
func (c *Control) UseTimer() {
	c.timer = time.Now()
}

// Body renders the given data into the response body
func (c *Control) Body(data interface{}) {
	var content []byte
	if str, ok := data.(string); ok {
		c.Writer.Header().Add("Content-type", MIMETEXT)
		content = []byte(str)
	} else {
		if !c.timer.IsZero() {
			took := time.Now()
			data = &Header{Took: took.Sub(c.timer), Data: data}
		}
		jsn, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		c.Writer.Header().Add("Content-type", MIMEJSON)
		content = jsn
	}
	if c.code > 0 {
		c.Writer.WriteHeader(c.code)
	}
	c.Writer.Write(content)
}
