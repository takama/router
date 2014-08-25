package main

import (
	"github.com/takama/router"
)

// Data is helper to construct JSON
type Data map[string]interface{}

func main() {
	r := router.New()
	r.GET("/hello/:name", func(c *router.Control) {
		c.Body("Hello " + c.Get(":name"))
	})

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
