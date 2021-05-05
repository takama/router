package router

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var r = New()

func TestRouterRegisterHandlers(t *testing.T) {

	// Create new Router
	//	r := New()

	// Registers GET handler for root static path
	r.GET("/", func(c *Control) {
		c.Body("Root")
	})

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

func TestRouterGetRootStatic(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Body.String() != "Root" {
		t.Error("Expected", "Root", "got", trw.Body.String())
	}
}

func TestRouterGetStatic(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Body.String() != "Hello" {
		t.Error("Expected", "Hello", "got", trw.Body.String())
	}
}

func TestRouterGetParameter(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello/John", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Body.String() != "Hello John" {
		t.Error("Expected", "Hello John", "got", trw.Body.String())
	}
}

func TestRouterGetParameterFromClassicUrl(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/Jane/?name=Joe", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Body.String() != "Users: Jane Joe" {
		t.Error("Expected", "Users: Jane Joe", "got", trw.Body.String())
	}
}

func TestRouterPostJSONData(t *testing.T) {
	req, err := http.NewRequest("POST", "/users/", strings.NewReader(`{"name": "Tom"}`))
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Body.String() != "User: Tom" {
		t.Error("Expected", "User: Tom", "got", trw.Body.String())
	}
}

func TestRouterPutJSONData(t *testing.T) {
	req, err := http.NewRequest("PUT", "/users/", strings.NewReader(`{"name1": "user1", "name2": "user2"}`))
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Body.String() != "Users: user1 user2" {
		t.Error("Expected", "Users: user1 user2", "got", trw.Body.String())
	}
}

func TestRouterDelete(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/users/", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Body.String() != "Users deleted" {
		t.Error("Expected", "Users deleted", "got", trw.Body.String())
	}
}

func TestRouterAllowedMethod(t *testing.T) {
	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Body.String() != "Method Not Allowed\n" {
		t.Error("Expected", "Method Not Allowed", "got", trw.Body.String())
	}
	methods := strings.Split(trw.Header().Get("Allow"), ", ")
	expected := "POST, PUT, DELETE"
	for _, method := range methods {
		if !strings.Contains("POST, PUT, DELETE", method) {
			t.Error("Expected", expected, "got", method)
		}
	}
}

func TestRouterNotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Code != http.StatusNotFound {
		t.Error("Expected", http.StatusNotFound, "got", trw.Code)
	}

}

func TestRouterPanic(t *testing.T) {
	req, err := http.NewRequest("GET", "/panic", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	r.ServeHTTP(trw, req)
	if trw.Code != http.StatusInternalServerError {
		t.Error("Expected", http.StatusInternalServerError, "got", trw.Code)
	}
}
