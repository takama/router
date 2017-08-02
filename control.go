// Copyright 2015 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Default content types
const (
	// MIMEJSON - "Content-type" for JSON
	MIMEJSON = "application/json"
	// MIMETEXT - "Content-type" for TEXT
	MIMETEXT = "text/plain"
)

// Control allows us to pass variables between middleware,
// assign Http codes and render a Body.
type Control struct {

	// Context embedded
	context.Context

	// Request is an adapter which allows the usage of a http.Request as standard request
	Request *http.Request

	// Writer is an adapter which allows the usage of a http.ResponseWriter as standard writer
	Writer http.ResponseWriter

	// User content type
	ContentType string

	// Code of HTTP status
	code int

	// compactJSON propery defines JSON output format (default is not compact)
	compactJSON bool

	// if used, json header shows meta data
	useMetaData bool

	// header with metadata
	header Header

	// errors
	errorHeader ErrorHeader

	// params is set of key/value parameters
	params []Param

	// timer used to calculate a elapsed time for handler and writing it in a response
	timer time.Time
}

// Param is a URL parameter which represents as key and value.
type Param struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// Header is used to prepare a JSON header with meta data
type Header struct {
	Duration   time.Duration `json:"duration,omitempty"`
	Took       string        `json:"took,omitempty"`
	APIVersion string        `json:"apiVersion,omitempty"`
	Context    string        `json:"context,omitempty"`
	ID         string        `json:"id,omitempty"`
	Method     string        `json:"method,omitempty"`
	Params     interface{}   `json:"params,omitempty"`
	Data       interface{}   `json:"data,omitempty"`
	Error      interface{}   `json:"error,omitempty"`
}

// ErrorHeader contains error code, message and array of specified error reports
type ErrorHeader struct {
	Code    uint16  `json:"code,omitempty"`
	Message string  `json:"message,omitempty"`
	Errors  []Error `json:"errors,omitempty"`
}

// Error report format
type Error struct {
	Domain       string `json:"domain,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	Location     string `json:"location,omitempty"`
	LocationType string `json:"locationType,omitempty"`
	ExtendedHelp string `json:"extendedHelp,omitempty"`
	SendReport   string `json:"sendReport,omitempty"`
}

// Get returns the first value associated with the given name.
// If there are no values associated with the key, an empty string is returned.
func (c *Control) Get(name string) string {
	for idx := range c.params {
		if c.params[idx].Key == name {
			return c.params[idx].Value
		}
	}

	return c.Request.URL.Query().Get(name)
}

// Set adds new parameters which represents as set of key/value.
func (c *Control) Set(params ...Param) *Control {
	c.params = append(c.params, params...)
	return c
}

// Code assigns http status code, which returns on http request
func (c *Control) Code(code int) *Control {
	if code >= 200 && code < 600 {
		c.code = code
	}
	return c
}

// GetCode returns status code
func (c *Control) GetCode() int {
	return c.code
}

// CompactJSON changes JSON output format (default mode is false)
func (c *Control) CompactJSON(mode bool) *Control {
	c.compactJSON = mode
	return c
}

// UseMetaData shows meta data in JSON Header
func (c *Control) UseMetaData() *Control {
	c.useMetaData = true
	return c
}

// APIVersion adds API version meta data
func (c *Control) APIVersion(version string) *Control {
	c.useMetaData = true
	c.header.APIVersion = version
	return c
}

// HeaderContext adds context meta data
func (c *Control) HeaderContext(context string) *Control {
	c.useMetaData = true
	c.header.Context = context
	return c
}

// ID adds id meta data
func (c *Control) ID(id string) *Control {
	c.useMetaData = true
	c.header.ID = id
	return c
}

// Method adds method meta data
func (c *Control) Method(method string) *Control {
	c.useMetaData = true
	c.header.Method = method
	return c
}

// SetParams adds params meta data in alternative format
func (c *Control) SetParams(params interface{}) *Control {
	c.useMetaData = true
	c.header.Params = params
	return c
}

// SetError sets error code and error message
func (c *Control) SetError(code uint16, message string) *Control {
	c.useMetaData = true
	c.errorHeader.Code = code
	c.errorHeader.Message = message
	return c
}

// AddError adds new error
func (c *Control) AddError(errors ...Error) *Control {
	c.useMetaData = true
	c.errorHeader.Errors = append(c.errorHeader.Errors, errors...)
	return c
}

// UseTimer allows caalculate elapsed time of request handling
func (c *Control) UseTimer() {
	c.useMetaData = true
	c.timer = time.Now()
}

// GetTimer returns timer state
func (c *Control) GetTimer() time.Time {
	return c.timer
}

// Body renders the given data into the response body
func (c *Control) Body(data interface{}) {
	var content []byte

	if str, ok := data.(string); ok {
		content = []byte(str)
		if c.ContentType != "" {
			c.Writer.Header().Add("Content-type", c.ContentType)
		} else {
			c.Writer.Header().Add("Content-type", MIMETEXT)
		}
	} else {
		if c.useMetaData {
			c.header.Data = data
			if !c.timer.IsZero() {
				took := time.Now()
				c.header.Duration = took.Sub(c.timer)
				c.header.Took = took.Sub(c.timer).String()
			}
			if c.header.Params == nil && len(c.params) > 0 {
				c.header.Params = c.params
			}
			if c.errorHeader.Code != 0 || c.errorHeader.Message != "" || len(c.errorHeader.Errors) > 0 {
				c.header.Error = c.errorHeader
			}
			data = c.header
		}
		var err error
		if c.compactJSON {
			content, err = json.Marshal(data)
		} else {
			content, err = json.MarshalIndent(data, "", "  ")
		}
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		c.Writer.Header().Add("Content-type", MIMEJSON)
	}
	if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
		c.Writer.Header().Add("Content-Encoding", "gzip")
		if c.code > 0 {
			c.Writer.WriteHeader(c.code)
		}
		gz := gzip.NewWriter(c.Writer)
		gz.Write(content)
		gz.Close()
	} else {
		if c.code > 0 {
			c.Writer.WriteHeader(c.code)
		}
		c.Writer.Write(content)
	}
}
