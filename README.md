Go Router
=========

A simple, compact and fast router package to process HTTP requests.
It has some sugar from framework however still lightweight. The router package is useful to prepare a RESTful API for Go services. It has JSON output, which bind automatically for relevant type of data. The router has timer feature to display duration of request handling in the header  

[![Build Status](https://travis-ci.org/takama/router.png?branch=master)](https://travis-ci.org/takama/router)
[![GoDoc](https://godoc.org/github.com/takama/router?status.svg)](https://godoc.org/github.com/takama/router)

### Examples

- Simplest example (serve static route): 
```go
package main

import (
	"github.com/takama/router"
)

func Hello(c *router.Control) {
	c.Body("Hello world")
}

func main() {
	r := router.New()
	r.GET("/hello", Hello)

	// Listen and serve on 0.0.0.0:8888
	r.Listen(":8888")
}
```

- Check it:
```sh
curl -i http://localhost:8888/hello/

HTTP/1.1 200 OK
Content-Type: text/plain
Date: Sun, 17 Aug 2014 13:25:50 GMT
Content-Length: 11

Hello world
```

- Serve dynamic route with parameter:
```go
package main

import (
	"github.com/takama/router"
)

func main() {
	r := router.New()
	r.GET("/hello/:name", func(c *router.Control) {
		c.Body("Hello " + c.Get(":name"))
	})

	// Listen and serve on 0.0.0.0:8888
	r.Listen(":8888")
}
```

- Check it:
```sh
curl -i http://localhost:8888/hello/John

HTTP/1.1 200 OK
Content-Type: text/plain
Date: Sun, 17 Aug 2014 13:25:55 GMT
Content-Length: 10

Hello John
```

- Checks JSON Content-Type automatically:
```go
package main

import (
	"github.com/takama/router"
)

// Data is helper to construct JSON
type Data map[string]interface{}

func main() {
	r := router.New()
	r.GET("/api/v1/settings/database/:db", func(c *router.Control) {
		data := Data{
			"Database settings": Data{
				"database": c.Get(":db"),
				"host":     "localhost",
				"port":     "3306",
			},
		}
		c.Code(200).Body(data)
	})
	// Listen and serve on 0.0.0.0:8888
	r.Listen(":8888")
}
```

- Check it:
```sh
curl -i http://localhost:8888/api/v1/settings/database/testdb

HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 17 Aug 2014 13:25:58 GMT
Content-Length: 102

{
  "Database settings": {
    "database": "testdb",
    "host": "localhost",
    "port": "3306"
  }
}
```

- Use timer to calculate duration of request handling:
```go
package main

import (
	"github.com/takama/router"
)

// Data is helper to construct JSON
type Data map[string]interface{}

func main() {
	r := router.New()
	r.GET("/api/v1/settings/database/:db", func(c *router.Control) {
		c.UseTimer()

		// Do something

		data := Data{
			"Database settings": Data{
				"database": c.Get(":db"),
				"host":     "localhost",
				"port":     "3306",
			},
		}
		c.Code(200).Body(data)
	})
	// Listen and serve on 0.0.0.0:8888
	r.Listen(":8888")
}
```

- Check it:
```sh
curl -i http://localhost:8888/api/v1/settings/database/testdb

HTTP/1.1 200 OK
Content-Type: application/json
Date: Sun, 17 Aug 2014 13:26:05 GMT
Content-Length: 143

{
  "duration": 5356123
  "took": "5.356ms",
  "data": {
    "Database settings": {
      "database": "testdb",
      "host": "localhost",
      "port": "3306"
    }
  }
}
```

- Custom handler with "Access-Control-Allow" options and compact JSON:
```go
package main

import (
	"github.com/takama/router"
)

// Data is helper to construct JSON
type Data map[string]interface{}

func baseHandler(handle router.Handle) router.Handle {
	return func(c *router.Control) {
		c.CompactJSON(true)
		if origin := c.Request.Header.Get("Origin"); origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		handle(c)
	}
}

func Info(c *router.Control) {
	data := Data{
		"debug": true,
		"error": false,
	}
	c.Body(data)
}

func main() {
	r := router.New()
	r.CustomHandler = baseHandler
	r.GET("/info", Info)

	// Listen and serve on 0.0.0.0:8888
	r.Listen(":8888")
}
```

- Check it:
```sh
curl -i -H 'Origin: http://foo.com' http://localhost:8888/info/

HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: http://foo.com
Content-Type: text/plain
Date: Sun, 17 Aug 2014 13:27:10 GMT
Content-Length: 28

{"debug":true,"error":false}
```

- Use google json style `https://google.github.io/styleguide/jsoncstyleguide.xml`:
```go
package main

import (
	"net/http"

	"github.com/takama/router"
)

func main() {
	r := router.New()
	r.GET("/api/v1/people/:action/:id", func(c *router.Control) {

		// Do something

		c.Method("people." + c.Get(":action"))
		c.SetParams(map[string]string{"userId": c.Get(":id")})
		c.SetError(http.StatusNotFound, "UserId not found")
		c.AddError(router.Error{Message: "Group or User not found"})
		c.Code(http.StatusNotFound).Body(nil)
	})
	// Listen and serve on 0.0.0.0:8888
	r.Listen(":8888")
}
```

- Check it:
```sh
curl -i http://localhost:8888/api/v1/people/get/@me

HTTP/1.1 404 Not Found
Content-Type: application/json
Date: Sat, 22 Oct 2016 14:50:00 GMT
Content-Length: 220

{
  "method": "people.get",
  "params": {
    "userId": "@me"
  },
  "error": {
    "code": 404,
    "message": "UserId not found",
    "errors": [
      {
        "message": "Group or User not found"
      }
    ]
  }
}
```

## Contributors (unsorted)

- [Igor Dolzhikov](https://github.com/takama)
- [Yaroslav Lukyanov](https://github.com/CSharpRU)

All the contributors are welcome. If you would like to be the contributor please accept some rules.
- The pull requests will be accepted only in "develop" branch
- All modifications or additions should be tested
- Sorry, I'll not accept code with any dependency, only standard library

Thank you for your understanding!

## License

[BSD License](https://github.com/takama/router/blob/master/LICENSE)
