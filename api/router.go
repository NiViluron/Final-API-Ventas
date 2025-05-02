package api

import (
	"Final-API-Ventas/internal/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitRoutes registers all user CRUD endpoints on the given GIN engine.
// It initializes the storage, service, and handler, then binds each HTTP
// method and path to the appropiate handler function.
func InitRoutes(e *gin.Engine) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	storage := user.NewLocalStorage()
	service := user.NewService(storage, logger)

	h := handler{
		userService: service,
		logger:      logger,
	}

	e.POST("/users", h.handleCreate)
	e.GET("/users/:id", h.handleRead)
	e.PATCH("/users/:id", h.handleUpdate)
	e.DELETE("/users/:id", h.handleDelete)

	e.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}
