package main

import (
	"flag"
	"net/http"

	"github.com/labstack/echo"
)

var (
	addr = flag.String("addr", "localhost:7001", "addr for backend")
)

func main() {
	flag.Parse()
	e := echo.New()
	// Routes
	e.GET("/hello", func(c echo.Context) error {
		value := map[string]interface{}{
			"message": "Hello, I am ReacllSong.",
		}
		return c.JSON(http.StatusOK, value)
	})
	e.GET("/hello/:id", func(c echo.Context) error {
		value := map[string]interface{}{
			"id":   c.Param("id"),
			"name": "ReacllSong",
		}
		return c.JSON(http.StatusOK, value)
	})
	e.GET("/hello/health", func(c echo.Context) error {
		value := map[string]interface{}{
			"status": "Ok",
		}
		return c.JSON(http.StatusOK, value)
	})

	// Start server
	e.Logger.Fatal(e.Start(*addr))
}
