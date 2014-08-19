// Copyright 2014 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestRouterRegisterHandlers(t *testing.T) {
	r := New()
	r.GET("/hello", func(c *Control) {
		c.Body("Hello")
	})
	r.GET("/hello/:name", func(c *Control) {
		c.Body("Hello " + c.Get(":name"))
	})
	r.GET("/users/:name", func(c *Control) {
		c.Body("Users: " + c.Get(":name") + " " + c.Get("name"))
	})
	r.POST("/users", func(c *Control) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			t.Error(err)
		}
		var values map[string]string
		if err := json.Unmarshal(body, &values); err != nil {
			t.Error(err)
		}
		c.Body("User: " + values["name"])
	})
	r.PUT("/users", func(c *Control) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			t.Error(err)
		}
		var values map[string]string
		if err := json.Unmarshal(body, &values); err != nil {
			t.Error(err)
		}
		c.Body("Users: " + values["name1"] + " " + values["name2"])
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
}

func TestRouterGetStatic(t *testing.T) {
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
}

func TestRouterGetParameter(t *testing.T) {
	response, err := http.Get("http://localhost:8888/hello/John")
	if err != nil {
		t.Error(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if string(body) != "Hello John" {
		t.Error("Expected", "Hello John", "got", string(body))
	}
}

func TestRouterGetParameterFromClassicUrl(t *testing.T) {
	response, err := http.Get("http://localhost:8888/users/Jane/?name=Joe")
	if err != nil {
		t.Error(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if string(body) != "Users: Jane Joe" {
		t.Error("Expected", "Users: Jane Joe", "got", string(body))
	}
}

func TestRouterPostJSONData(t *testing.T) {
	reader := strings.NewReader(`{"name": "Tom"}`)
	response, err := http.Post("http://localhost:8888/users/", MIMEJSON, reader)
	if err != nil {
		t.Error(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if string(body) != "User: Tom" {
		t.Error("Expected", "User: Tom", "got", string(body))
	}

}

func TestRouterPutJSONData(t *testing.T) {
	reader := strings.NewReader(`{"name1": "user1", "name2": "user2"}`)
	client := new(http.Client)
	req, err := http.NewRequest("PUT", "http://localhost:8888/users/", reader)
	if err != nil {
		t.Error(err)
	}
	response, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if string(body) != "Users: user1 user2" {
		t.Error("Expected", "Users: user1 user2", "got", string(body))
	}
}
