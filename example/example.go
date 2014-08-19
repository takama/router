package main

import (
	"github.com/takama/router"
)

func main() {
	r := router.New()
	r.GET("/hello/:name", func(c *router.Control) {
		c.Body("Hello " + c.Get(":name"))
	})

	r.GET("/api/v1/settings/database/:db", func(c *router.Control) {
		data := map[string]map[string]string{
			"Database settings": {
				"database": c.Get(":db"),
				"host":     "localhost",
				"port":     "3306",
			},
		}
		c.UseTimer()
		c.Code(200).Body(data)
	})
	// Listen and serve on 0.0.0.0:8888
	r.Listen(":8888")
}
