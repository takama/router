package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type registered struct {
	path   string
	handle Handle
}

type expected struct {
	data       string
	request    string
	route      string
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
	{
		"/static/*",
		func(c *Control) {
			c.Body("Hello from star static path")
		},
	},
	{
		"/files/:dir/*",
		func(c *Control) {
			c.Body(c.Get(":dir"))
		},
	},
}

var setOfExpected = []expected{
	{
		"Hello Jane",
		"/hello/Jane",
		"/hello/:name",
		1,
		[]Param{
			{":name", "Jane"},
		},
	},
	{
		"Hello from static path",
		"/hello/John",
		"/hello/John",
		0,
		[]Param{},
	},
	{
		"jack from hell",
		"/hell/jack",
		"/:h/:n",
		2,
		[]Param{
			{":h", "hell"},
			{":n", "jack"},
		},
	},
	{
		"Product: table order# 23",
		"/products/table/orders/23",
		"/products/:name/orders/:id",
		2,
		[]Param{
			{":name", "table"},
			{":id", "23"},
		},
	},
	{
		"Product: book order# 12",
		"/products/book/orders/12",
		"/products/book/orders/:id",
		1,
		[]Param{
			{":id", "12"},
		},
	},
	{
		"Product: pen order# 11",
		"/products/pen/orders/11",
		"/products/:name/orders/:id",
		2,
		[]Param{
			{":name", "pen"},
			{":id", "11"},
		},
	},
	{
		"Product: pen # 10",
		"/products/pen/order/10",
		"/products/:name/:order/:id",
		3,
		[]Param{
			{":name", "pen"},
			{":order", "order"},
			{":id", "10"},
		},
	},
	{
		"product pen order # 10",
		"/product/pen/order/10",
		"/:product/:name/:order/:id",
		4,
		[]Param{
			{":product", "product"},
			{":name", "pen"},
			{":order", "order"},
			{":id", "10"},
		},
	},
	{
		"Hello from star static path",
		"/static/greetings/something",
		"/static/*",
		0,
		[]Param{},
	},
	{
		"css",
		"/files/css/style.css",
		"/files/:dir/*",
		1,
		[]Param{
			{":dir", "css"},
		},
	},
	{
		"js",
		"/files/js/app.js",
		"/files/:dir/*",
		1,
		[]Param{
			{":dir", "js"},
		},
	},
}

func TestParserRegisterGet(t *testing.T) {
	p := newParser()
	for _, request := range setOfRegistered {
		p.register(request.path, request.handle)
	}
	for _, exp := range setOfExpected {
		handle, params, route, ok := p.get(exp.request)
		if !ok {
			t.Error("Error: get data for path", exp.request)
		}
		if len(params) != exp.paramCount {
			t.Error("Expected length of param", exp.paramCount, ", got", len(params))
		}
		if route != exp.route {
			t.Error("Expected route", exp.route, ", got", route)
		}
		c := new(Control)
		c.Set(params...)
		trw := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "", nil)
		if err != nil {
			t.Error("Error creating new request")
		}
		c.Writer, c.Request = trw, req
		handle(c)
		if trw.Body.String() != exp.data {
			t.Error("Expected", exp.data, ", got", trw.Body.String())
		}
	}
}

func TestParserSplit(t *testing.T) {
	path := []string{
		"/api/v1/module",
		"/api//v1/module/",
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

	if part, ok := split("   "); ok {
		if len(part) != 0 {
			t.Error("Error: split data for path '/'", part)
		}
	} else {
		t.Error("Error: split data for path '/'")
	}

	if part, ok := split("///"); ok {
		if len(part) != 0 {
			t.Error("Error: split data for path '/'", part)
		}
	} else {
		t.Error("Error: split data for path '/'")
	}

	if part, ok := split("  /  //  "); ok {
		if len(part) != 0 {
			t.Error("Error: split data for path '/'", part)
		}
	} else {
		t.Error("Error: split data for path '/'")
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
