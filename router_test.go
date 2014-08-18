// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestRouterSimplestUsing(t *testing.T) {
	r := New()
	r.GET("/hello", func(c *Control) {
		c.Body("Hello")
	})
	r.GET("/hello/:name", func(c *Control) {
		c.Body("Hello " + c.Get(":name"))
	})
	go r.Listen(":8888")
	if handle, params, ok := r.Lookup("GET", "/hello"); ok {
		if handle == nil {
			t.Error("Handle not initialized")
		}
		if len(params) != 0 {
			t.Error("Expected len of params", 0, "got", len(params))
		}
	} else {
		t.Error("Path not fouund: /hello")
	}
	if handle, params, ok := r.Lookup("GET", "/hello/John"); ok {
		if handle == nil {
			t.Error("Handle not initialized")
		}
		if len(params) != 1 {
			t.Error("Expected len of params", 1, "got", len(params))
		}
		if params[0].Key != ":name" {
			t.Error("Expected key", ":name", "got", params[0].Key)
		}
		if params[0].Value != "John" {
			t.Error("Expected value", "Jonn", "got", params[0].Key)
		}
	} else {
		t.Error("Path not fouund: /hello/John")
	}
	response, err := http.Get("http://localhost:8888/hello")
	if err != nil {
		t.Error(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if string(body) != "Hello" {
		t.Error("Expected", "Hello", "got", string(body))
	}
	response, err = http.Get("http://localhost:8888/hello/John")
	if err != nil {
		t.Error(err)
	}
	body, err = ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if string(body) != "Hello John" {
		t.Error("Expected", "Hello John", "got", string(body))
	}
}
