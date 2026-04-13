package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/AarambhDevHub/rudra/core"
	rudraContext "github.com/AarambhDevHub/rudra/context"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var (
	users  = make(map[int64]*User)
	nextID atomic.Int64
)

func main() {
	app := core.New()

	// Seed some data
	nextID.Store(1)
	createUser("Arjun", 28)
	createUser("Priya", 24)

	// API v1 group
	api := app.Group("/api/v1")

	// Users group
	usersGrp := api.Group("/users")
	usersGrp.GET("", listUsers)
	usersGrp.GET("/:id", getUser)
	usersGrp.POST("", createUserHandler)
	usersGrp.DELETE("/:id", deleteUser)

	log.Println("rudra: REST API starting on :8080")
	if err := app.Run(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatalf("rudra: server error: %v", err)
	}
}

func listUsers(c *rudraContext.Context) error {
	list := make([]*User, 0, len(users))
	for _, u := range users {
		list = append(list, u)
	}
	return c.JSON(http.StatusOK, list)
}

func getUser(c *rudraContext.Context) error {
	id := c.Param("id")
	for _, u := range users {
		if matchID(u, id) {
			return c.JSON(http.StatusOK, u)
		}
	}
	return c.AbortWithError(http.StatusNotFound, http.ErrAbortHandler)
}

func createUserHandler(c *rudraContext.Context) error {
	var input struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := c.BindJSON(&input); err != nil {
		return c.AbortWithError(http.StatusBadRequest, err)
	}
	u := createUser(input.Name, input.Age)
	return c.JSON(http.StatusCreated, u)
}

func deleteUser(c *rudraContext.Context) error {
	id := c.Param("id")
	for k, u := range users {
		if matchID(u, id) {
			delete(users, k)
			return c.NoContent()
		}
	}
	return c.AbortWithError(http.StatusNotFound, http.ErrAbortHandler)
}

func createUser(name string, age int) *User {
	id := nextID.Add(1) - 1
	u := &User{ID: id, Name: name, Age: age}
	users[id] = u
	return u
}

func matchID(u *User, id string) bool {
	return fmt.Sprintf("%d", u.ID) == id
}