package main

import (
	"log"

	"MyBlog/internal/handlers"
	"MyBlog/internal/store"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Static("/", "./web")

	store := store.NewStore()
	handler := handlers.New(store)
	handler.RegisterRoutes(r)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
