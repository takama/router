// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"testing"
)

var parameters = []Param{
	{"name", "John"},
	{"age", "32"},
	{"gender", "M"},
}

type testResponseWriter struct {
	code int
	data []byte
}

func (trw *testResponseWriter) Header() http.Header {
	return http.Header{}
}

func (trw *testResponseWriter) Write(data []byte) (int, error) {
	trw.data = data
	return len(data), nil
}

func (trw *testResponseWriter) WriteHeader(code int) {
	trw.code = code
}

func TestControlSetGet(t *testing.T) {

	c := new(Control)
	c.Set(parameters)
	for _, param := range parameters {
		if c.Get(param.Key) != param.Value {
			t.Error("Expected for", param.Key, ":", param.Value, ", got", c.Get(param.Key))
		}
	}
}

func TestControlCode(t *testing.T) {
	c := new(Control)
	// code transcends, must be less than 600
	c.Code(606)
	if c.code != 0 {
		t.Error("Expected code", "0", "got", c.code)
	}
	c.Code(404)
	if c.code != 404 {
		t.Error("Expected code", "404", "got", c.code)
	}
}

func TestControlBody(t *testing.T) {
	trw := new(testResponseWriter)
	req, err := http.NewRequest("GET", "hello/:name", nil)
	if err != nil {
		t.Error("Error creting new request")
	}
	c := new(Control)
	c.Writer, c.Request = trw, req
	c.Body("Hello")
	if string(trw.data) != "Hello" {
		t.Error("Expected", "Hello", "got", string(trw.data))
	}
	c.Body(parameters)
	if string(trw.data) != testJSONData {
		t.Error("Expected", testJSONData, "got", string(trw.data))
	}
}

var testJSONData = `[
  {
    "Key": "name",
    "Value": "John"
  },
  {
    "Key": "age",
    "Value": "32"
  },
  {
    "Key": "gender",
    "Value": "M"
  }
]`
