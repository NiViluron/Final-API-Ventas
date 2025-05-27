package main

import (
	"Final-API-Ventas/api"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	api.InitRoutes(r, "http://localhost:8080")

	if err := r.Run(":8081"); err != nil {
		panic(fmt.Errorf("error trying to start server: %v", err))
	}
}
