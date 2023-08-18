package main

import (
	// "fmt"
	// "log"
	// "net/http"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	g := gin.Default();

	g.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the backend")
	})

	g.Run(":8080")
}