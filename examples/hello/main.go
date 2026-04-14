package main

import (
	"log"
	"net/http"

	rudraContext "github.com/AarambhDevHub/rudra/context"
	"github.com/AarambhDevHub/rudra/core"
)

func main() {
	app := core.New()

	app.GET("/", func(c *rudraContext.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"framework": "Rudra",
			"status":    "fierce",
		})
	})

	app.GET("/hello/:name", func(c *rudraContext.Context) error {
		name := c.Param("name")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Hello, " + name + "!",
		})
	})

	go func() {
		log.Println("rudra: starting server on :8080")
		if err := app.Run(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("rudra: server error: %v", err)
		}
	}()

	app.ListenForShutdown()
}
