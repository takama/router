package router

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestRouterRegisterHandlers(t *testing.T) {

	// Create new Router
	r := New()

	// Registers GET handler for static path
	r.GET("/hello", func(c *Control) {
		c.Body("Hello")
	})

	// Registers GET handler with parameter
	r.GET("/hello/:name", func(c *Control) {
		c.Body("Hello " + c.Get(":name"))
	})

	// Registers GET handler with two parameters
	r.GET("/users/:name", func(c *Control) {
		c.Body("Users: " + c.Get(":name") + " " + c.Get("name"))
	})

	// Registers POST handler
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

	// Registers PUT handler
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

	// Registers DELETE handler
	r.DELETE("/users", func(c *Control) {
		c.Body("Users deleted")
	})

	// Registers thrown panic method and panic handler
	r.GET("/panic", func(c *Control) {
		panic("Thrown panic")
	})
	r.PanicHandler = func(c *Control) {
		c.Code(http.StatusInternalServerError).Body("Internal Server Error")
	}

	// Listen and serve handlers
	go r.Listen(":8888")

	// Checks parameters for static path method
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

	// Checks parameters for dynamic path method
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

func TestRouterDelete(t *testing.T) {
	client := new(http.Client)
	req, err := http.NewRequest("DELETE", "http://localhost:8888/users/", nil)
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
	if string(body) != "Users deleted" {
		t.Error("Users deleted", "got", string(body))
	}
}

func TestRouterAllowedMethod(t *testing.T) {
	response, err := http.Get("http://localhost:8888/users")
	if err != nil {
		t.Error(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if string(body) != "Method Not Allowed\n" {
		t.Error("Expected", "Method Not Allowed", "got", string(body))
	}
	methods := strings.Split(response.Header.Get("Allow"), ", ")
	expected := "POST, PUT, DELETE"
	for _, method := range methods {
		if !strings.Contains("POST, PUT, DELETE", method) {
			t.Error("Expected", expected, "got", method)
		}
	}
}

func TestRouterNotFound(t *testing.T) {
	response, err := http.Get("http://localhost:8888/test")
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Error("Expected", http.StatusNotFound, "got", response.StatusCode)
	}

}

func TestRouterPanic(t *testing.T) {
	response, err := http.Get("http://localhost:8888/panic")
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != http.StatusInternalServerError {
		t.Error("Expected", http.StatusInternalServerError, "got", response.StatusCode)
	}
}
