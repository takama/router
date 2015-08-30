// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var parameters = []Param{
	{"name", "John"},
	{"age", "32"},
	{"gender", "M"},
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
	req, err := http.NewRequest("GET", "hello/:name", nil)
	if err != nil {
		t.Error("Error creting new request")
	}
	c := new(Control)
	trw := httptest.NewRecorder()
	c.Writer, c.Request = trw, req
	c.Body("Hello")
	if trw.Body.String() != "Hello" {
		t.Error("Expected", "Hello", "got", trw.Body.String())
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	c.Body(123)
	if trw.Body.String() != "123" {
		t.Error("Expected", "123", "got", trw.Body.String())
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	c.Body(123.1)
	if trw.Body.String() != "123.1" {
		t.Error("Expected", "123.1", "got", trw.Body.String())
	}
	trw = httptest.NewRecorder()
	c.Writer = trw
	c.Body(parameters)
	if trw.Body.String() != testJSONData {
		t.Error("Expected", testJSONData, "got", trw.Body.String())
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
