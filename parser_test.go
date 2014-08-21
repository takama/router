// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"net/http"
	"testing"
)

type registered struct {
	path   string
	handle Handle
}

type expected struct {
	request    string
	data       string
	paramCount int
	params     []Param
}

var setOfRegistered = []registered{
	{
		"/hello/:name",
		func(c *Control) {
			c.Body("Hello " + c.Get(":name"))
		},
	},
	{
		"/hello/John",
		func(c *Control) {
			c.Body("Hello from static path")
		},
	},
	{
		"/:h/:n",
		func(c *Control) {
			c.Body(c.Get(":n") + " from " + c.Get(":h"))
		},
	},
	{
		"/products/:name/orders/:id",
		func(c *Control) {
			c.Body("Product: " + c.Get(":name") + " order# " + c.Get(":id"))
		},
	},
	{
		"/products/book/orders/:id",
		func(c *Control) {
			c.Body("Product: book order# " + c.Get(":id"))
		},
	},
	{
		"/products/:name/:order/:id",
		func(c *Control) {
			c.Body("Product: " + c.Get(":name") + " # " + c.Get(":id"))
		},
	},
	{
		"/:product/:name/:order/:id",
		func(c *Control) {
			c.Body(c.Get(":product") + " " + c.Get(":name") + " " + c.Get(":order") + " # " + c.Get(":id"))
		},
	},
}

var setOfExpected = []expected{
	{
		"/hello/Jane",
		"Hello Jane",
		1,
		[]Param{
			{":name", "Jane"},
		},
	},
	{
		"/hello/John",
		"Hello from static path",
		0,
		[]Param{},
	},
	{
		"/hell/jack",
		"jack from hell",
		2,
		[]Param{
			{":h", "hell"},
			{":n", "jack"},
		},
	},
	{
		"/products/table/orders/23",
		"Product: table order# 23",
		2,
		[]Param{
			{":name", "table"},
			{":id", "23"},
		},
	},
	{
		"/products/book/orders/12",
		"Product: book order# 12",
		1,
		[]Param{
			{":id", "12"},
		},
	},
	{
		"/products/pen/orders/11",
		"Product: pen order# 11",
		2,
		[]Param{
			{":name", "pen"},
			{":id", "11"},
		},
	},
	{
		"/products/pen/order/10",
		"Product: pen # 10",
		3,
		[]Param{
			{":name", "pen"},
			{":order", "order"},
			{":id", "10"},
		},
	},
	{
		"/product/pen/order/10",
		"product pen order # 10",
		4,
		[]Param{
			{":product", "product"},
			{":name", "pen"},
			{":order", "order"},
			{":id", "10"},
		},
	},
}

func TestParserRegisterGet(t *testing.T) {
	p := newParser()
	for _, request := range setOfRegistered {
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
		req, err := http.NewRequest("GET", "", nil)
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
