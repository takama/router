// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"testing"
)

type expected struct {
	path       string
	request    string
	handle     Handle
	data       string
	paramCount int
	params     []Param
}

var setOfExpected = []expected{
	{
		"/hello/:name",
		"/hello/Jane",
		func(c *Control) {
			c.Body("Hello " + c.Get(":name"))
		},
		"Hello Jane",
		1,
		[]Param{
			{":name", "Jane"},
		},
	},
	{
		"/hello/John",
		"/hello/John",
		func(c *Control) {
			c.Body("Hello from static path")
		},
		"Hello from static path",
		0,
		[]Param{},
	},
	{
		"/:h/:n",
		"/hell/jack",
		func(c *Control) {
			c.Body(c.Get(":n") + " from " + c.Get(":h"))
		},
		"jack from hell",
		2,
		[]Param{
			{":h", "hell"},
			{":n", "jack"},
		},
	},
}

func TestParserRegisterGet(t *testing.T) {
	p := newParser()
	for _, request := range setOfExpected {
		p.register(request.path, request.handle)
	}
	for _, exp := range setOfExpected {
		handle, params, ok := p.get(exp.request)
		if !ok {
			t.Error("Error: get data for path", exp.request)
		}
		if len(params) != exp.paramCount {
			t.Error("Expected length of param", exp.paramCount, "got", len(params))
		}
		c := new(Control)
		c.Set(params)
		trw := new(testResponseWriter)
		req, err := http.NewRequest("GET", exp.path, nil)
		if err != nil {
			t.Error("Error creating new request")
		}
		c.Writer, c.Request = trw, req
		handle(c)
		if string(trw.data) != exp.data {
			t.Error("Expected", exp.data, "got", string(trw.data))
		}
	}
}

func TestParserSplit(t *testing.T) {
	path := []string{
		"/api/v1/module",
		"/api/v1/module/",
		"/module///name//",
		"module//:name",
		"/:param1/:param2/",
	}
	expected := [][]string{
		{"api", "v1", "module"},
		{"api", "v1", "module"},
		{"module", "name"},
		{"module", ":name"},
		{":param1", ":param2"},
	}
	for idx, p := range path {
		parts, ok := split(p)
		if !ok {
			t.Error("Error: split data for path", p)
		}
		for i, part := range parts {
			if expected[idx][i] != part {
				t.Error("Expected", expected[idx][i], "got", part)
			}
		}
	}
}
